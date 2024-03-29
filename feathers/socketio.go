package feathers

import (
	"errors"
	"net/http"
	"reflect"
	"sync"

	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
	gosocketio "github.com/tobiasbeck/feathers-go/gosf-socketio"
	"github.com/tobiasbeck/feathers-go/gosf-socketio/transport"
)

func stringToCallmethod(method string) RestMethod {
	switch method {
	case "create":
		return Create
	case "update":
		return Update
	case "patch":
		return Patch
	case "remove":
		return Remove
	case "find":
		return Find
	case "get":
		return Get
	}
	panic("unknown server method")
}

type socketConnection struct {
	channel    *gosocketio.Channel
	authEntity interface{}
}

func (c *socketConnection) Join(room string) error {
	return c.channel.Join(room)
}
func (c *socketConnection) Leave(room string) error {
	return c.channel.Leave(room)
}

func (c *socketConnection) AuthEntity() interface{} {
	return c.authEntity
}

func (c *socketConnection) SetAuthEntity(entity interface{}) {
	if c.authEntity != nil {
		return
	}
	c.authEntity = entity
}

func (c *socketConnection) Emit(event string, data interface{}) error {
	return c.channel.Emit(event, data)
}

func (c *socketConnection) IsAuthenticated() bool {
	return c.authEntity != nil
}

type socketCaller struct {
	connection    *socketConnection
	channel       *gosocketio.Channel
	response      chan<- interface{}
	errorResponse chan<- error
}

func (c *socketCaller) Callback(data interface{}) {
	c.response <- data
	close(c.response)
}

func (c *socketCaller) CallbackError(data error) {
	c.errorResponse <- data
}

func (c *socketCaller) IsSocket() bool {
	return true
}

func (c *socketCaller) SocketConnection() Connection {
	return c.connection
}

//SocketIOProvider handles socket.io connections and events
type SocketIOProvider struct {
	server      *gosocketio.Server
	app         *App
	connections map[string]*socketConnection
}

//Publish publishes a event to connections subscibed to room
func (p *SocketIOProvider) Publish(room string, event string, data interface{}, path string, provider string) {
	p.server.BroadcastTo(room, event, data)
}

// Creates a new SocketIOProvider instance (use module `ConfigureSocketIOProvider` with apps `Configure` method)
func NewSocketIOProvider(app *App, config map[string]interface{}) *SocketIOProvider {
	provider := new(SocketIOProvider)
	provider.connections = make(map[string]*socketConnection)
	provider.server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	provider.app = app
	provider.listenEvent("create")
	provider.listenEvent("update")
	provider.listenEvent("patch")
	provider.listenEvent("remove")
	provider.listenEvent("find")
	provider.listenEvent("get")
	provider.server.On(gosocketio.OnConnection, func(channel *gosocketio.Channel) {
		connection := &socketConnection{
			channel: channel,
		}
		provider.connections[channel.Id()] = connection
		provider.app.Emit("connection", channel)
	})
	provider.server.On(gosocketio.OnDisconnection, func(channel *gosocketio.Channel) {
		if socketchannel, ok := provider.connections[channel.Id()]; ok {
			delete(provider.connections, channel.Id())
			provider.app.Emit("disconnect", socketchannel)
		}

	})
	return provider
}

// ConfigureSocketIOProvider registers a new socketio provider in app
func ConfigureSocketIOProvider(app *App, config map[string]interface{}) error {
	return app.AddProvider("socketio", NewSocketIOProvider(app, config))
}

func (p *SocketIOProvider) listenEvent(event string) {
	p.server.On(event, func(c *gosocketio.Channel, data []interface{}) interface{} {
		response := make(chan interface{}, 0)
		responseError := make(chan error, 0)
		p.handleEvent(event, c, response, responseError, data)
		select {
		case data := <-response:
			return data
		case err := <-responseError:
			return err
		}
	})
}

//Handle handles a new event to a service
func (p *SocketIOProvider) Handle(callMethod RestMethod, caller socketCaller, service string, data map[string]interface{}, id string, query map[string]interface{}) {
	p.app.HandleRequest("socketio", callMethod, &caller, service, data, id, query)
}

// Listen starts listening for new socket.io connections
func (fs *SocketIOProvider) Listen(port int, serveMux *http.ServeMux) {
	serveMux.Handle("/socket.io/", fs.server)
}

func (fs *SocketIOProvider) handleEvent(event string, c *gosocketio.Channel, response chan<- interface{}, responseErr chan<- error, data []interface{}) {
	if len(data) <= 0 {
		go func() {
			responseErr <- httperrors.NewBadRequest("Service not defined")
		}()
		return
	}
	serviceType := reflect.TypeOf(data[0])
	if serviceType.String() != "string" {
		go func() {
			responseErr <- httperrors.NewBadRequest("Service name not string")
		}()
		return
	}

	connection, ok := fs.connections[c.Id()]

	if !ok {
		go func() {
			responseErr <- httperrors.NewBadRequest("Connection is not registered")
		}()
		return
	}

	service := data[0].(string)
	caller := socketCaller{
		channel:       c,
		connection:    connection,
		response:      response,
		errorResponse: responseErr,
	}

	callMethod := stringToCallmethod(event)
	var reqData map[string]interface{} = map[string]interface{}{}
	reqQuery := make(map[string]interface{})
	var id string = ""
	if len(data) >= 2 {
		switch v := data[1].(type) {
		case string:
			id = data[1].(string)
			if len(data) >= 3 {
				if secondData, ok := data[2].(map[string]interface{}); ok {
					reqQuery = filterData(dl_query, callMethod, secondData)
					reqData = filterData(dl_data, callMethod, secondData)
				}
			}
		case nil:
			id = ""
			if len(data) >= 3 {
				if secondData, ok := data[2].(map[string]interface{}); ok {
					reqQuery = filterData(dl_query, callMethod, secondData)
					reqData = filterData(dl_data, callMethod, secondData)
				}
			}
		case map[string]interface{}:
			reqQuery = filterData(dl_query, callMethod, v)
			reqData = filterData(dl_data, callMethod, v)
		}
	}
	if len(data) >= 4 {
		reqQuery = data[3].(map[string]interface{})
	}
	fs.Handle(callMethod, caller, service, reqData, id, reqQuery)
}

// PublishHandler is a function which handles a publish of a service and returns a list of rooms to publish to
type PublishHandler = func(data interface{}, ctx *Context) []string

// PublishableService which can publish events
type PublishableService interface {
	RegisterPublishHandler(topic string, handler PublishHandler)
	Publish(topic string, data interface{}, ctx *Context) ([]string, error)
	BeforePublish(topic string, data interface{}, ctx *Context) (interface{}, error)
}

//BasePublishableService is a basic implementation of PublishableService
type BasePublishableService struct {
	events     map[string]PublishHandler
	eventsLock sync.RWMutex
}

func NewBasePublishableService() *BasePublishableService {
	return &BasePublishableService{
		events: map[string]PublishHandler{},
	}
}

// RegisterPublishHandler registers a new handler for a topic
func (s *BasePublishableService) RegisterPublishHandler(topic string, handler PublishHandler) {
	s.eventsLock.Lock()
	defer s.eventsLock.Unlock()
	s.events[topic] = handler
}

// Publish calls PublishHandler if registerd and publishes data to returned topics
func (s *BasePublishableService) Publish(topic string, data interface{}, ctx *Context) ([]string, error) {
	s.eventsLock.RLock()
	defer s.eventsLock.RUnlock()
	if handler, ok := s.events[topic]; ok {
		result := handler(data, ctx)
		return result, nil
	}
	return nil, errors.New("Handler is not registered")
}

// BeforePublish is called before data is published to a topic. data can be manipulated at this point
func (s *BasePublishableService) BeforePublish(topic string, data interface{}, ctx *Context) (interface{}, error) {
	return data, nil
}
