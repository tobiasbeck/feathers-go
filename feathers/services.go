package feathers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator"
	"github.com/mcuadros/go-defaults"
)

func mergeHooks(chainA []Hook, chainB []Hook) []Hook {

	copyA := make([]func(ctx *Context) error, len(chainA))
	copyB := make([]func(ctx *Context) error, len(chainB))
	copy(copyA, chainA)
	copy(copyB, chainB)
	return append(copyA, copyB...)
}

//Service  is a callable instance inside feathers which is responslible for a single kind of entity
type Service interface {
	// Find retrieves multiple entities (`interface{} is a slice`)
	Find(ctx context.Context, params Params) (interface{}, error)
	// Get retrives a single entity
	Get(ctx context.Context, id string, params Params) (interface{}, error)
	// Create creates a new entity should be created
	Create(ctx context.Context, data map[string]interface{}, params Params) (interface{}, error)
	// Update replaces a whole entity
	Update(ctx context.Context, id string, data map[string]interface{}, params Params) (interface{}, error)
	// Patch updates or replaces specified entity keys
	Patch(ctx context.Context, id string, data map[string]interface{}, params Params) (interface{}, error)
	// Remove removes a entity
	Remove(ctx context.Context, id string, params Params) (interface{}, error)

	// HookTree returns the hook tree (mainly uses internally)
	HookTree() HooksTree

	// Name returns the name of the current service
	Name() string

	// setName sets a service name (used internally)
	setName(name string)
}

// HooksTreeBranch is a single branch of hooks (e.g. for Before, After or Error)
type HooksTreeBranch struct {
	// All hooks are executed every time before route specific hooks are executed
	All []Hook
	// Find hooks are executed for find calls
	Find []Hook
	// Get hooks are executed for get calls
	Get []Hook
	// Create hooks are executed for create calls
	Create []Hook
	// Patch hooks are executed for patch calls
	Patch []Hook
	// Update hooks are executed for update calls
	Update []Hook
	// Remove hooks are executed for remove hooks
	Remove []Hook
}

func (b HooksTreeBranch) Branch(method RestMethod) []Hook {
	key := strings.Title(method.String())
	if chain, ok := getField(&b, key); ok {
		hc := chain.([]Hook)
		return mergeHooks(b.All, hc)
	}
	return make([]Hook, 0)
}

func (b *HooksTreeBranch) Append(method RestMethod, hooks []Hook) {
	key := strings.Title(method.String())
	if chain, ok := getField(&b, key); ok {
		hc := chain.([]Hook)
		merged := mergeHooks(hc, hooks)
		setField(b, key, merged)
	} else {
		panic(fmt.Sprintf("Could not find branch %s", key))
	}
}

//HooksTree is the complete hooks definition of a service or the application
type HooksTree struct {
	//Before hooks are executed before service method
	Before HooksTreeBranch
	// After hooks are executed after service method
	After HooksTreeBranch
	// Error hooks are executed in case hook or service method returns error
	Error HooksTreeBranch
}

func (t HooksTree) Branch(branchType HookType) HooksTreeBranch {
	key := strings.Title(branchType.String())
	if branch, ok := getField(&t, key); ok {
		hc := branch.(HooksTreeBranch)
		return hc
	}
	panic("unknown hook tree")
}

func (t *HooksTree) Append(branchType HookType, method RestMethod, hooks []Hook) {
	key := strings.Title(branchType.String())
	if branch, ok := getField(&t, key); ok {
		hc := branch.(HooksTreeBranch)
		hc.Append(method, hooks)
	} else {
		panic(fmt.Sprintf("Could not find branch %s", key))
	}
}

type Defaulter interface {
	SetDefaults()
}

// BaseService (every service should extend from this)
type BaseService struct {
	Hooks HooksTree
	name  string
}

func (b *BaseService) Name() string {
	return b.name
}

func (b *BaseService) setName(name string) {
	b.name = name
}

// HookTree returns hook tree of service
func (b *BaseService) HookTree() HooksTree {
	return b.Hooks
}

type Mappable interface {
	ToMap() map[string]interface{}
}

// ModelFactory returns a new instance of the model (used for `ModelService` struct)
type ModelFactory = func() interface{}

//ModelService is a service which offers model validation and parsing (create new with `NewModelService`)
type ModelService struct {
	Model     ModelFactory
	validator *validator.Validate
}

// MapToModel parses data passed to a service and returns a model instance
func (m *ModelService) MapToModel(data map[string]interface{}) (interface{}, error) {
	model := m.Model()
	decoder, err := newDecoder(model)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	if defaulter, ok := model.(Defaulter); ok {
		defaulter.SetDefaults()
	} else {
		// TODO: this is a deprecated fallback and will be removed in the future
		defaults.SetDefaults(model)
	}

	return model, nil
}

// StructToMap converts a model struct into an interface
func (m *ModelService) StructToMap(data interface{}) (map[string]interface{}, error) {
	if mappableStruct, ok := data.(Mappable); ok {
		return mappableStruct.ToMap(), nil
	}
	result := make(map[string]interface{})
	decoder, err := newDecoder(&result)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MapAndValidate is the same as calling `MapToModel` and `ValidateModel`
func (m *ModelService) MapAndValidate(data map[string]interface{}) (interface{}, error) {
	model, err := m.MapToModel(data)
	if err != nil {
		return nil, err
	}
	err = m.ValidateModel(model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// MapAndValidate is the same as calling `MapToStruct` and `ValidateModel` on the returned struct
func (m *ModelService) MapAndValidateStruct(data map[string]interface{}, target interface{}) error {
	err := MapToStruct(data, target)
	if err != nil {
		return err
	}
	err = m.ValidateModel(target)
	if err != nil {
		return err
	}
	return nil
}

// ValidateModel validates a model based on its validation rules
func (m *ModelService) ValidateModel(model interface{}) error {
	err := m.validator.Struct(model)
	return err
}

// Creates a new ModelService based of a existing model
func NewModelService(model ModelFactory) *ModelService {
	return &ModelService{
		Model:     model,
		validator: validator.New(),
	}
}

type appServiceCaller struct {
	success chan interface{}
	err     chan error
}

func (asc *appServiceCaller) Callback(data interface{}) {
	asc.success <- data
}
func (asc *appServiceCaller) CallbackError(err error) {
	asc.err <- err
}

func (c *appServiceCaller) IsSocket() bool {
	return false
}

func (c *appServiceCaller) SocketConnection() Connection {
	return nil
}

type appService struct {
	app     *App
	service Service
	name    string
}

func (as *appService) Find(ctx context.Context, params Params) (interface{}, error) {

	return as.callMethod(ctx, Find, map[string]interface{}{}, "", params)
}

func (as *appService) Get(ctx context.Context, id string, params Params) (interface{}, error) {
	return as.callMethod(ctx, Get, map[string]interface{}{}, id, params)
}

func (as *appService) Create(ctx context.Context, data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(ctx, Create, data, "", params)
}

func (as *appService) Update(ctx context.Context, id string, data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(ctx, Update, data, id, params)
}

func (as *appService) Patch(ctx context.Context, id string, data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(ctx, Patch, data, id, params)
}

func (as *appService) Remove(ctx context.Context, id string, params Params) (interface{}, error) {
	return as.callMethod(ctx, Remove, map[string]interface{}{}, id, params)
}

func (as *appService) HookTree() HooksTree {

	return as.service.HookTree()
}

func (as *appService) Name() string {
	return as.service.Name()
}

func (as *appService) setName(name string) {
	// Does nothing
}

// RegisterPublishHandler registers a new handler for a topic
func (s *appService) RegisterPublishHandler(topic string, handler PublishHandler) {
	if ps, ok := s.service.(PublishableService); ok {
		ps.RegisterPublishHandler(topic, handler)
	}
}

// Publish calls PublishHandler if registerd and publishes data to returned topics
func (s *appService) Publish(topic string, data interface{}, ctx *Context) ([]string, error) {
	if ps, ok := s.service.(PublishableService); ok {
		result, err := ps.Publish(topic, data, ctx)
		return result, err
	}
	return []string{}, nil
}

// BeforePublish is called before data is published to a topic. data can be manipulated at this point
func (s *appService) BeforePublish(topic string, data interface{}, ctx *Context) (interface{}, error) {
	if ps, ok := s.service.(PublishableService); ok {
		result, err := ps.BeforePublish(topic, data, ctx)
		return result, err
	}
	return data, nil
}

func (as *appService) callMethod(ctx context.Context, method RestMethod, data map[string]interface{}, id string, params Params) (interface{}, error) {
	caller := &appServiceCaller{
		success: make(chan interface{}, 0),
		err:     make(chan error, 0),
	}

	as.app.handleServerServiceCall(ctx, as.name, method, caller, data, id, params)
	select {
	case result := <-caller.success:
		return result, nil
	case err := <-caller.err:
		return nil, err
	}
}
