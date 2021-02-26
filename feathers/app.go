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
	method    CallMethod
	chainType HookType
	hookChain []func(ctx *HookContext) (*HookContext, error) // If this is changed to []Hook it triggers #25838 in Go.
	service   Service
	position  int
	context   *HookContext
}

func paramContextCancelled(ctx HookContext) bool {
	context := ctx.Params.CallContext
	if context.Err() == nil {
		return false
	}
	return true
}

type Provider interface {
	Listen(port int, mux *http.ServeMux)
	Publish(room string, event string, data interface{}, provider string)
}

type AppModule = func(app *App, config map[string]interface{}) error

type App struct {
	*EventEmitter
	providers map[string]Provider

	hooks HooksTree

	services map[string]Service

	tasks chan task

	servicesLock sync.RWMutex

	config map[string]interface{}
}

// Add a network provider to go-feathers app
func (a *App) AddProvider(name string, provider Provider) error {
	a.providers[name] = provider
	return nil
}

func (a *App) Configure(appModule AppModule, config map[string]interface{}) {
	err := appModule(a, config)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) AddService(name string, service Service) {
	a.servicesLock.Lock()
	defer a.servicesLock.Unlock()
	a.services[name] = service
}

func (a *App) Startup(executorSize int) {
	for i := 0; i < executorSize; i++ {
		go a.workTask()
	}
}

func (a *App) HandleRequest(provider string, method CallMethod, c Caller, service string, data interface{}, id string) {
	if serviceInstance, ok := a.services[service]; ok {
		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			<-context.Done()
		}()
		initContext := HookContext{
			App:     *a,
			Data:    data,
			Method:  method,
			ID:      id,
			Service: serviceInstance,
			Type:    Before,
			Params: HookParams{
				Params:            make(map[string]interface{}),
				Provider:          provider,
				Route:             service,
				CallContext:       context,
				CallContextCancel: cancel,
				Headers:           "",
				fields:            make(map[string]interface{}),
			},
		}
		// fmt.Printf("Handle: Hooks %#v", serviceInstance.GetHooks())
		a.scheduleTask(tm_hook, c, Before, serviceInstance.GetHooks().Before.GetBranch(method), serviceInstance, 0, &initContext)
		return
	}
	fmt.Println("Unknown Service " + service)
	return
}

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
		fmt.Printf("Method returned error %#s\n", err.Error())
		errorContext := serviceTask.context
		errorContext.Type = Error
		errorContext.Error = err
		a.scheduleTask(tm_hook, serviceTask.caller, Error, serviceTask.service.GetHooks().Error.GetBranch(serviceTask.method), serviceTask.service, 0, errorContext)
		return
	}
	afterContext := serviceTask.context
	afterContext.Type = After
	afterContext.Result = result
	a.scheduleTask(tm_hook, serviceTask.caller, After, serviceTask.service.GetHooks().After.GetBranch(serviceTask.method), serviceTask.service, 0, afterContext)
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
			hookTask.context.Params.CallContextCancel()
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
			hookTask.context.Params.CallContextCancel()
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
		a.scheduleTask(tm_hook, hookTask.caller, Error, hookTask.service.GetHooks().Error.GetBranch(hookTask.method), hookTask.service, 0, errorContext)
		return
	}
	a.scheduleTask(tm_hook, hookTask.caller, hookTask.chainType, hookTask.hookChain, hookTask.service, hookTask.position+1, context)

}

func (a *App) mergeAppHooks(chain []func(ctx *HookContext) (*HookContext, error), hookType HookType, branch CallMethod) []func(ctx *HookContext) (*HookContext, error) {
	if appHooks, ok := getField(&a.hooks, strings.Title(hookType.String())); ok {
		appHookBranch := appHooks.(HooksTreeBranch)
		appHookChain := appHookBranch.GetBranch(branch)
		chainCopy := make([]func(ctx *HookContext) (*HookContext, error), len(chain))
		appHooksCopy := make([]func(ctx *HookContext) (*HookContext, error), len(appHookChain))
		copy(appHooksCopy, appHookChain)
		copy(chainCopy, chain)
		return append(appHooksCopy, chainCopy...)
	}
	return chain
}

func (a *App) PublishToProviders(room string, event string, data interface{}, publishProvider string) {
	for _, provider := range a.providers {
		provider.Publish(room, event, data, publishProvider)
	}
}

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

func (a *App) LoadConfig(path string) error {
	if config, err := loadConfig(path); err == nil {
		a.config = config
		return nil
	} else {
		return err
	}
}

func (a *App) SetAppHooks(hookTree HooksTree) {
	a.hooks = hookTree
}

func (a *App) GetConfig(key string) (interface{}, bool) {
	value, ok := a.config[key]
	return value, ok
}

func (a *App) SetConfig(key string, value interface{}) {
	a.config[key] = value
}

func (a *App) Service(name string) Service {
	if service, ok := a.services[name]; ok {
		return service
	}
	return nil
}

func (a *App) ServiceClass(name string) (interface{}, error) {
	if service, ok := a.services[name]; ok {
		return service, nil
	}
	return nil, errors.New("Service does not exist")
}

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
