package feathers

import (
	"errors"
	"net/http"
	"reflect"
	"sync"

	gosocketio "github.com/tobiasbeck/feathers-go/gosf-socketio"
	"github.com/tobiasbeck/feathers-go/gosf-socketio/transport"
)

func getCallMethod(method string) CallMethod {
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

type socketCaller struct {
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
	close(c.errorResponse)
}

type SocketIOProvider struct {
	server *gosocketio.Server
	app    *App
}

func (p *SocketIOProvider) Publish(room string, event string, data interface{}, provider string) {
	p.server.BroadcastTo(room, event, data)
}

func NewSocketIOProvider(app *App, config map[string]interface{}) *SocketIOProvider {
	provider := new(SocketIOProvider)
	provider.server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	provider.app = app
	provider.listenEvent("create")
	provider.listenEvent("update")
	provider.listenEvent("patch")
	provider.listenEvent("remove")
	provider.listenEvent("find")
	provider.listenEvent("get")
	provider.server.On(gosocketio.OnConnection, func(channel interface{}) {
		provider.app.Emit("connection", channel)
	})
	return provider
}

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

func (p *SocketIOProvider) Handle(callMethod CallMethod, caller socketCaller, service string, data map[string]interface{}, id string, query map[string]interface{}) {
	p.app.HandleRequest("socketio", callMethod, &caller, service, data, id, query)
}

func (fs *SocketIOProvider) Listen(port int, serveMux *http.ServeMux) {
	serveMux.Handle("/socket.io/", fs.server)
}

func (fs *SocketIOProvider) handleEvent(event string, c *gosocketio.Channel, response chan<- interface{}, responseErr chan<- error, data []interface{}) {
	serviceType := reflect.TypeOf(data[0])
	if serviceType.String() != "string" {
		return
	}
	service := data[0].(string)
	caller := socketCaller{
		channel:       c,
		response:      response,
		errorResponse: responseErr,
	}

	callMethod := getCallMethod(event)
	var reqData map[string]interface{}
	var reqQuery map[string]interface{}
	var id string

	switch v := data[1].(type) {
	case string:
		id = data[1].(string)
		if secondData, ok := data[2].(map[string]interface{}); ok {
			reqQuery = filterData(dl_query, callMethod, secondData)
			reqData = filterData(dl_data, callMethod, secondData)
		}
	case map[string]interface{}:
		reqQuery = filterData(dl_query, callMethod, v)
		reqData = filterData(dl_data, callMethod, v)
	}
	if len(data) >= 4 {
		reqQuery = data[3].(map[string]interface{})
	}
	fs.Handle(callMethod, caller, service, reqData, id, reqQuery)
}

// Publishable Service
type PublishHandler = func(data interface{}, ctx HookContext) []string

type PublishableService interface {
	RegisterPublishHandler(topic string, handler PublishHandler)
	Publish(topic string, data interface{}, ctx HookContext) ([]string, error)
}

type BasePublishableService struct {
	events     map[string]PublishHandler
	eventsLock sync.RWMutex
}

func (s *BasePublishableService) RegisterPublishHandler(topic string, handler PublishHandler) {
	s.eventsLock.Lock()
	defer s.eventsLock.Unlock()
	s.events[topic] = handler
}

func (s *BasePublishableService) Publish(topic string, data interface{}, ctx HookContext) ([]string, error) {
	s.eventsLock.RLock()
	defer s.eventsLock.RUnlock()
	if handler, ok := s.events[topic]; ok == true {
		return handler(data, ctx), nil
	}
	return nil, errors.New("Handler is not registered")
}
