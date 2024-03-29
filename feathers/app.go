package feathers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
)

type taskMode string

const (
	tm_hook    taskMode = "hook"
	tm_chain   taskMode = "chain"
	tm_service taskMode = "service"
)

type task struct {
	caller    Caller
	mode      taskMode
	method    RestMethod
	chainType HookType
	hookChain []func(ctx *Context) error // If this is changed to []Hook it triggers #25838 in Go.
	service   Service
	position  int
	context   *Context
}

// ---------------

type locationType string

const (
	dl_query locationType = "query"
	dl_data  locationType = "data"
)

func filterLocation(dataLocation locationType, location locationType, data map[string]interface{}) map[string]interface{} {
	if dataLocation == location {
		return data
	}
	return make(map[string]interface{})
}

func filterData(location locationType, method RestMethod, data map[string]interface{}) map[string]interface{} {
	switch method {
	case Find:
		return filterLocation(dl_query, location, data)
	case Get:
		return filterLocation(dl_query, location, data)
	case Create:
		return filterLocation(dl_data, location, data)
	case Remove:
		return filterLocation(dl_query, location, data)
	case Patch:
		return filterLocation(dl_data, location, data)
	case Update:
		return filterLocation(dl_data, location, data)
	}
	return nil
}

func paramContextCancelled(ctx Context) bool {
	context := ctx.Context
	if context.Err() == nil {
		return false
	}
	return true
}

// ---------------

// Provider handles requests and can listen for new connections
type Provider interface {
	Listen(port int, mux *http.ServeMux)
	Publish(room string, event string, data interface{}, path string, provider string)
}

type Setupable interface {
	Setup(app *App)
}

// AppModules is a module which can configure the application
/*
Are supposed to be passed to Configure method of App
*/
type AppModule = func(app *App, config map[string]interface{}) error

// App is the feathers-go applications instance. Instanciate through NewApp
type App struct {
	*EventEmitter
	providers map[string]Provider

	hooks HooksTree

	services map[string]Service

	servicesLock sync.RWMutex

	server *http.Server

	config map[string]interface{}
}

// ---------------

func (a *App) mergeAppHooks(chain []func(ctx *Context) error, hookType HookType, branch RestMethod) []func(ctx *Context) error {
	if appHooks, ok := getField(&a.hooks, strings.Title(hookType.String())); ok {
		appHookBranch := appHooks.(HooksTreeBranch)
		appHookChain := appHookBranch.Branch(branch)
		return mergeHooks(appHookChain, chain)
	}
	return chain
}

// ---------------

// AddProvider adds a network provider to the app.
/*
Primarily called through modules
Providers can listen for client connections and are also handling requests
*/
func (a *App) AddProvider(name string, provider Provider) error {
	a.providers[name] = provider
	return nil
}

// Configure configures modules similar to the original feathers api.
/*
Modules can registers Services, Providers etc.
*/
func (a *App) Configure(appModule AppModule, config map[string]interface{}) {
	err := appModule(a, config)
	if err != nil {
		log.Fatal(err)
	}
}

// AddService registers a new service for the application
func (a *App) AddService(name string, service Service) {
	a.servicesLock.Lock()
	defer a.servicesLock.Unlock()
	a.services[name] = service
	service.setName(name)
}

// Startup setups execution pool of go routines for pipeline execution
// func (a *App) Startup(executorSize int) {
// 	for i := 0; i < executorSize; i++ {
// 		go a.workTask()
// 	}
// }

func (a *App) handleServerServiceCall(ctx context.Context, service string, method RestMethod, c Caller, data interface{}, id string, params Params) {
	if serviceInstance, ok := a.services[service]; ok {
		initContext := Context{
			Context:      ctx,
			App:          *a,
			Data:         data.(map[string]interface{}),
			Method:       method,
			Path:         service,
			ID:           id,
			Service:      a.Service(service),
			ServiceClass: serviceInstance,
			Type:         Before,
			Params:       params,
		}
		go a.handlePipeline(&initContext, serviceInstance, c)
		return
	}
	c.CallbackError(httperrors.NewNotFound(fmt.Sprintf("Unknown Service %s", service)))
	log.Warnln("Unknown Service " + service)
	return
}

// HandleRequest handles a request received by a provider. It starts the pipeline and schedules tasks
func (a *App) HandleRequest(provider string, method RestMethod, c Caller, service string, data map[string]interface{}, id string, query map[string]interface{}) {
	// fmt.Printf("Request:\n  service: %s\n  method: %s\n  data: %+v\n query: %+v\n\n", service, method, data, query)
	if serviceInstance, ok := a.services[service]; ok {
		context, _ := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			<-context.Done()
		}()

		authenticated := false

		if connection := c.SocketConnection(); connection != nil {
			authenticated = connection.IsAuthenticated()
		}

		var user map[string]interface{}
		if connection := c.SocketConnection(); connection != nil {
			if connection.IsAuthenticated() {
				user = connection.AuthEntity().(map[string]interface{})
			}

		}

		initContext := Context{
			Context:      context,
			App:          *a,
			Data:         data,
			Method:       method,
			Path:         service,
			ID:           id,
			Service:      a.Service(service),
			ServiceClass: serviceInstance,
			Type:         Before,
			Params: Params{
				Params:        make(map[string]interface{}),
				Provider:      provider,
				Route:         service,
				Connection:    c.SocketConnection(),
				IsSocket:      c.IsSocket(),
				User:          user,
				Headers:       make(map[string]string),
				fields:        make(map[string]interface{}),
				Query:         query,
				Authenticated: authenticated,
			},
		}
		go a.handlePipeline(&initContext, serviceInstance, c)
		return
	}
	go func() {
		log.Warnln("Unknown Service" + service)
		c.CallbackError(httperrors.NewNotFound(fmt.Sprintf("Unknown Service %s", service)))
	}()
	return
}

func (a *App) handlePipeline(ctx *Context, service Service, c Caller) {
	var err error

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		c.CallbackError(r.(error))
	// 	}
	// }()

	// Before
	origCtx := ctx
	ctx, err = a.handleHookChain(ctx, Before, service)
	if err != nil {
		a.handlePipelineError(err, origCtx, service, c)
		return
	}
	if ctx.Result == nil {
		var result interface{}
		switch ctx.Method {
		case Create:
			result, err = service.Create(ctx, ctx.Data, ctx.Params)
		case Update:
			result, err = service.Update(ctx, ctx.ID, ctx.Data, ctx.Params)
		case Patch:
			result, err = service.Patch(ctx, ctx.ID, ctx.Data, ctx.Params)
		case Remove:
			result, err = service.Remove(ctx, ctx.ID, ctx.Params)
		case Find:
			result, err = service.Find(ctx, ctx.Params)
		case Get:
			result, err = service.Get(ctx, ctx.ID, ctx.Params)
		}
		if err != nil {
			a.handlePipelineError(err, ctx, service, c)
			return
		}
		ctx.Result = result
	}
	ctx, err = a.handleHookChain(ctx, After, service)
	if err != nil {
		a.handlePipelineError(err, origCtx, service, c)
		return
	}
	c.Callback(ctx.Result)
	go a.TriggerUpdate(ctx)

}

func (a *App) TriggerUpdate(ctx *Context) {
	// fmt.Printf("TRIGGER UPDATE: %s\n\n", ctx.Path)
	// fmt.Printf("Service: %T\n", ctx.ServiceClass)

	//Afterwards trigger updates
	if service, ok := ctx.Service.(PublishableService); ok {
		if event := eventFromCallMethod(ctx.Method); event != "" {
			if rooms, err := service.Publish(event, ctx.Result, ctx); err == nil {
				serviceEvent := fmt.Sprintf("%s %s", ctx.Path, event)
				for _, room := range rooms {
					data, err := service.BeforePublish(room, ctx.Result, ctx)
					if err != nil || data == nil {
						fmt.Println("SKIP SENDING", err, data)
						continue
					}
					a.PublishToProviders(room, serviceEvent, data, ctx.Path, ctx.Params.Provider)
				}
			}
		}
	}
}

func (a *App) handleHookChain(ctx *Context, chainType HookType, service Service) (*Context, error) {
	tree := service.HookTree()
	branch := tree.Branch(chainType)
	chain := branch.Branch(ctx.Method)
	mergedChain := a.mergeAppHooks(chain, chainType, ctx.Method)
	ctx.Type = chainType
	loopCtx := ctx
	for _, hook := range mergedChain {
		err := hook(loopCtx)
		if err != nil {
			return nil, err
		}
	}

	return loopCtx, nil
}

func (a *App) handlePipelineError(err error, ctx *Context, service Service, c Caller) {
	featherError := err
	if _, ok := err.(httperrors.FeathersError); !ok {
		featherError = httperrors.Convert(err)
	}
	ctx.Error = featherError
	ctx, chainErr := a.handleHookChain(ctx, Error, service)
	if chainErr != nil {
		c.CallbackError(chainErr)
		return
	}
	c.CallbackError(ctx.Error)
}

// PublishToProviders publishes a event to all providers. Each can decide what to do with the publish by themselfs
func (a *App) PublishToProviders(room string, event string, data interface{}, path string, publishProvider string) {
	for _, provider := range a.providers {
		provider.Publish(room, event, data, path, publishProvider)
	}
}

// Listen makes the app listen to the port specified in the configuration
func (a *App) Listen() {
	if len(a.providers) == 0 {
		panic("No providers configured")
	}

	if port, ok := a.config["port"]; ok {
		log.Infoln("Listening at ", port)
		// mux := mux.NewRouter()

		for _, service := range a.services {
			setupable, ok := service.(Setupable)
			if !ok {
				continue
			}
			setupable.Setup(a)
		}

		serveMux := http.NewServeMux()
		for _, provider := range a.providers {
			provider.Listen(port.(int), serveMux)
		}
		a.server = &http.Server{Addr: ":" + strconv.Itoa(port.(int)), Handler: serveMux}
		err := a.server.ListenAndServe()
		if err != nil {
			log.Panic(err)
		}
		// log.Panic(http.ListenAndServe(":"+strconv.Itoa(port.(int)), serveMux))
		return
	}
	log.Fatal("Could not find port (may not specified in config)")
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.server == nil {
		return errors.New("Server not running")
	}
	return a.server.Shutdown(ctx)
}

func (a *App) Close() error {
	if a.server == nil {
		return errors.New("Server not running")
	}
	return a.server.Close()
}

//LoadConfig loads configuration files from `./config directory`
/* By default `default.yaml` file is loaded.
It also looks for a environment variable named `APP_ENV`.
If it is specified it looks for a config file named after the environment and merges it with default configuration
Configuration can be retrieved by `GetConfig`. Keys kan be set and overwritten by `SetConfig`
Example:
In case of `APP_ENV=development`, looks for `development.yaml` and if it exists merges it with APP_ENV
*/
func (a *App) LoadConfig() error {
	if config, err := loadConfig("./config"); err == nil {
		a.config = config
		return nil
	} else {
		return err
	}
}

// SetAppHooks specifies the HookTree of the application (Similar to feathers-go)
func (a *App) SetAppHooks(hookTree HooksTree) {
	a.hooks = hookTree
}

// Config retrieves a config key from the applications config
func (a *App) Config(key string) (interface{}, bool) {
	value, ok := a.config[key]
	return value, ok
}

// SetConfig sets a config key in the applications config
func (a *App) SetConfig(key string, value interface{}) {
	a.config[key] = value
}

// Service returns a a wrapped service instance of a service for easy calling methods
/*
Wrapping is necessary because otherwise hooks will not be triggered
If service does not exist returns  nil
*/
func (a *App) Service(name string) Service {
	if service, ok := a.services[name]; ok {
		return &appService{
			app:     a,
			name:    name,
			service: service,
		}
	}
	return nil
}

// ServiceClass returns a service instance as an interface{}
/*
This is useful for parsing a service to a specific interface or struct for calling custom service methods
*/
func (a *App) ServiceClass(name string) interface{} {
	if service, ok := a.services[name]; ok {
		return service
	}
	return nil
}

// NewApp returns a new feathers-go app instance
func NewApp() *App {
	app := &App{
		EventEmitter: NewEventEmitter(),
		providers:    make(map[string]Provider, 0),
		services:     make(map[string]Service, 0),
		hooks:        HooksTree{},
		config:       make(map[string]interface{}),
	}
	return app
}
