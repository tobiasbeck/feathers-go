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

	tasks chan task

	servicesLock sync.RWMutex

	config map[string]interface{}
}

// ---------------

func (a *App) workTask() {
	for doTask := range a.tasks {
		if paramContextCancelled(*doTask.context) {
			return
		}
		switch doTask.mode {
		case tm_hook:
			a.processHook(doTask)
		case tm_service:
			a.processServiceMethod(doTask)
		}
	}
}

func (a *App) processServiceMethod(serviceTask task) {
	// serviceTask.context.Params.CallContext.
	var result interface{}
	var err error
	switch serviceTask.method {
	case Create:
		result, err = serviceTask.service.Create(serviceTask.context.Data.(map[string]interface{}), serviceTask.context.Params)
	case Update:
		result, err = serviceTask.service.Update(serviceTask.context.ID, serviceTask.context.Data.(map[string]interface{}), serviceTask.context.Params)
	case Patch:
		result, err = serviceTask.service.Patch(serviceTask.context.ID, serviceTask.context.Data.(map[string]interface{}), serviceTask.context.Params)
	case Remove:
		result, err = serviceTask.service.Remove(serviceTask.context.ID, serviceTask.context.Params)
	case Find:
		result, err = serviceTask.service.Find(serviceTask.context.Params)
	case Get:
		result, err = serviceTask.service.Get(serviceTask.context.ID, serviceTask.context.Params)
	}
	if err != nil {
		// fmt.Printf("Method returned error: %#s\n", err.Error())
		errorContext := serviceTask.context
		errorContext.Type = Error
		errorContext.Error = err
		a.scheduleTask(tm_hook, serviceTask.caller, Error, serviceTask.service.HookTree().Error.Branch(serviceTask.method), serviceTask.service, 0, errorContext)
		return
	}
	afterContext := serviceTask.context
	afterContext.Type = After
	afterContext.Result = result
	a.scheduleTask(tm_hook, serviceTask.caller, After, serviceTask.service.HookTree().After.Branch(serviceTask.method), serviceTask.service, 0, afterContext)
}

func (a *App) scheduleTask(mode taskMode, caller Caller, chainType HookType, hookChain []func(ctx *HookContext) (*HookContext, error), service Service, position int, context *HookContext) {
	var mergedChain []func(ctx *HookContext) (*HookContext, error)
	if position == 0 {
		mergedChain = a.mergeAppHooks(hookChain, context.Type, context.Method)
	} else {
		mergedChain = hookChain
	}
	a.tasks <- task{
		mode:      mode,
		method:    context.Method,
		caller:    caller,
		chainType: chainType,
		hookChain: mergedChain,
		service:   service,
		position:  position,
		context:   context,
	}
}

func (a *App) processHook(hookTask task) {
	// fmt.Printf("processHook: Type: %s, Mode: %s, Chain Len: %s, Position: %s\n", hookTask.chainType, hookTask.mode, len(hookTask.hookChain), hookTask.position)
	if len(hookTask.hookChain) == 0 || len(hookTask.hookChain) <= hookTask.position {
		switch hookTask.chainType {
		case Before:
			a.scheduleTask(tm_service, hookTask.caller, Before, nil, hookTask.service, 0, hookTask.context)
		case After:
			hookTask.caller.Callback(hookTask.context.Result)
			if hookTask.context.Params.CallContextCancel != nil {
				hookTask.context.Params.CallContextCancel()
			}
			if service, ok := hookTask.context.Service.(PublishableService); ok {
				if event := eventFromCallMethod(hookTask.context.Method); event != "" {
					if rooms, err := service.Publish(event, hookTask.context.Result, *hookTask.context); err != nil {
						for _, room := range rooms {
							a.PublishToProviders(room, event, hookTask.context.Result, hookTask.context.Params.Provider)
						}
					}
				}
			}
		case Error:
			if hookTask.context.Params.CallContextCancel != nil {
				hookTask.context.Params.CallContextCancel()
			}
			hookTask.caller.CallbackError(hookTask.context.Error)
		}
		return
	}
	hook := hookTask.hookChain[hookTask.position]
	context, err := hook(hookTask.context)
	if err != nil {
		errorContext := context
		if errorContext == nil {
			errorContext = hookTask.context
		}
		errorContext.Type = Error
		errorContext.Error = err
		a.scheduleTask(tm_hook, hookTask.caller, Error, hookTask.service.HookTree().Error.Branch(hookTask.method), hookTask.service, 0, errorContext)
		return
	}
	a.scheduleTask(tm_hook, hookTask.caller, hookTask.chainType, hookTask.hookChain, hookTask.service, hookTask.position+1, context)

}

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
func (a *App) Startup(executorSize int) {
	for i := 0; i < executorSize; i++ {
		go a.workTask()
	}
}

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
		a.scheduleTask(tm_hook, c, Before, serviceInstance.HookTree().Before.Branch(method), serviceInstance, 0, &initContext)
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
				CallContextCancel: cancel,
				Headers:           make(map[string]string),
				fields:            make(map[string]interface{}),
				Query:             query,
			},
		}
		a.scheduleTask(tm_hook, c, Before, serviceInstance.HookTree().Before.Branch(method), serviceInstance, 0, &initContext)
		return
	}
	fmt.Println("Unknown Service " + service)
	return
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
		tasks:        make(chan task, 500),
		hooks:        HooksTree{},
		config:       make(map[string]interface{}),
	}
	return app
}
