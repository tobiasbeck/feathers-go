package feathers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
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
	hookChain []func(ctx *HookContext) (*HookContext, error) // If this is changed to []Hook it triggers #25838 in Go.
	service   Service
	position  int
	context   *HookContext
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

func paramContextCancelled(ctx HookContext) bool {
	context := ctx.Params.CallContext
	if context.Err() == nil {
		return false
	}
	return true
}

// ---------------

// Provider handles requests and can listen for new connections
type Provider interface {
	Listen(port int, mux *http.ServeMux)
	Publish(room string, event string, data interface{}, provider string)
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

	config map[string]interface{}
}

// ---------------

func (a *App) mergeAppHooks(chain []func(ctx *HookContext) (*HookContext, error), hookType HookType, branch RestMethod) []func(ctx *HookContext) (*HookContext, error) {
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
}

// Startup setups execution pool of go routines for pipeline execution
// func (a *App) Startup(executorSize int) {
// 	for i := 0; i < executorSize; i++ {
// 		go a.workTask()
// 	}
// }

func (a *App) handleServerServiceCall(service string, method RestMethod, c Caller, data interface{}, id string, params Params) {
	if serviceInstance, ok := a.services[service]; ok {
		initContext := HookContext{
			App:     *a,
			Data:    data,
			Method:  method,
			Path:    service,
			ID:      id,
			Service: serviceInstance,
			Type:    Before,
			Params:  params,
		}
		go a.handlePipeline(&initContext, serviceInstance, c)
		return
	}
	fmt.Println("Unknown Service " + service)
	return
}

// HandleRequest handles a request received by a provider. It starts the pipeline and schedules tasks
func (a *App) HandleRequest(provider string, method RestMethod, c Caller, service string, data interface{}, id string, query map[string]interface{}) {
	if serviceInstance, ok := a.services[service]; ok {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

		initContext := HookContext{
			App:     *a,
			Data:    data,
			Method:  method,
			Path:    service,
			ID:      id,
			Service: serviceInstance,
			Type:    Before,
			Params: Params{
				Params:            make(map[string]interface{}),
				Provider:          provider,
				Route:             service,
				CallContext:       context,
				Connection:        c.SocketConnection(),
				IsSocket:          c.IsSocket(),
				User:              user,
				CallContextCancel: cancel,
				Headers:           make(map[string]string),
				fields:            make(map[string]interface{}),
				Query:             query,
				Authenticated:     authenticated,
			},
		}
		go a.handlePipeline(&initContext, serviceInstance, c)
		return
	}
	fmt.Println("Unknown Service " + service)
	return
}

func (a *App) handlePipeline(ctx *HookContext, service Service, c Caller) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Printf("GOT PANIC %#v\n", err)
	// 		//c.CallbackError(err.(error))
	// 	}
	// }()
	var err error
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
			result, err = service.Create(ctx.Data.(map[string]interface{}), ctx.Params)
		case Update:
			result, err = service.Update(ctx.ID, ctx.Data.(map[string]interface{}), ctx.Params)
		case Patch:
			result, err = service.Patch(ctx.ID, ctx.Data.(map[string]interface{}), ctx.Params)
		case Remove:
			result, err = service.Remove(ctx.ID, ctx.Params)
		case Find:
			result, err = service.Find(ctx.Params)
		case Get:
			result, err = service.Get(ctx.ID, ctx.Params)
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
}

func (a *App) handleHookChain(ctx *HookContext, chainType HookType, service Service) (*HookContext, error) {
	tree := service.HookTree()
	branch := tree.Branch(chainType)
	chain := branch.Branch(ctx.Method)
	mergedChain := a.mergeAppHooks(chain, chainType, ctx.Method)
	ctx.Type = chainType
	loopCtx := ctx
	for _, hook := range mergedChain {
		result, err := hook(loopCtx)
		if err != nil {
			return nil, err
		}
		loopCtx = result
	}

	return loopCtx, nil
}

func (a *App) handlePipelineError(err error, ctx *HookContext, service Service, c Caller) {
	ctx.Error = err
	ctx, chainErr := a.handleHookChain(ctx, Error, service)
	if chainErr != nil {
		c.CallbackError(chainErr)
		return
	}
	c.CallbackError(err)
}

// PublishToProviders publishes a event to all providers. Each can decide what to do with the publish by themselfs
func (a *App) PublishToProviders(room string, event string, data interface{}, publishProvider string) {
	for _, provider := range a.providers {
		provider.Publish(room, event, data, publishProvider)
	}
}

// Listen makes the app listen to the port specified in the configuration
func (a *App) Listen() {
	if len(a.providers) == 0 {
		panic("No providers configured")
	}

	if port, ok := a.config["port"]; ok {
		fmt.Println("Listening at ", port)
		// mux := mux.NewRouter()
		serveMux := http.NewServeMux()
		for _, provider := range a.providers {
			provider.Listen(port.(int), serveMux)
		}
		log.Println("Listening...")
		log.Panic(http.ListenAndServe(":"+strconv.Itoa(port.(int)), serveMux))
		return
	}
	log.Fatal("Could not find port (may not specified in config)")
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
func (a *App) ServiceClass(name string) (interface{}, error) {
	if service, ok := a.services[name]; ok {
		return service, nil
	}
	return nil, errors.New("Service does not exist")
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
