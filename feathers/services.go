package feathers

import (
	"strings"

	"github.com/go-playground/validator"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
)

func mergeHooks(chainA []Hook, chainB []Hook) []Hook {

	copyA := make([]func(ctx *HookContext) (*HookContext, error), len(chainA))
	copyB := make([]func(ctx *HookContext) (*HookContext, error), len(chainB))
	copy(copyA, chainA)
	copy(copyB, chainB)
	return append(copyA, copyB...)
}

//Service  is a callable instance inside feathers which is responslible for a single kind of entity
type Service interface {
	// Find retrieves multiple entities (`interface{} is a slice`)
	Find(params Params) (interface{}, error)
	// Get retrives a single entity
	Get(id string, params Params) (interface{}, error)
	// Create creates a new entity should be created
	Create(data map[string]interface{}, params Params) (interface{}, error)
	// Update replaces a whole entity
	Update(id string, data map[string]interface{}, params Params) (interface{}, error)
	// Patch updates or replaces specified entity keys
	Patch(id string, data map[string]interface{}, params Params) (interface{}, error)
	// Remove removes a entity
	Remove(id string, params Params) (interface{}, error)

	//HookTree returns the hook tree (mainly uses internally)
	HookTree() HooksTree
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
	if chain, ok := getField(&b, key); ok == true {
		hc := chain.([]Hook)
		return mergeHooks(b.All, hc)
	}
	return make([]Hook, 0)
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
	if branch, ok := getField(&t, key); ok == true {
		hc := branch.(HooksTreeBranch)
		return hc
	}
	panic("unknown hook tree")
}

// BaseService (every service should extend from this)
type BaseService struct {
	Hooks HooksTree
}

// HookTree returns hook tree of service
func (b *BaseService) HookTree() HooksTree {
	return b.Hooks
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
	err := mapstructure.Decode(data, model)
	if err != nil {
		return nil, err
	}
	defaults.SetDefaults(model)
	return model, nil
}

// StructToMap converts a model struct into an interface
func (m *ModelService) StructToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := mapstructure.Decode(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MapToStruct maps service data into a struct (passed by pointer)
/*
Example:
````
model := Model{}
err := MapToStruct(data, &model)
````
*/
func MapToStruct(data map[string]interface{}, target interface{}) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	defaults.SetDefaults(target)
	return nil
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

func (as *appService) Find(params Params) (interface{}, error) {

	return as.callMethod(Find, map[string]interface{}{}, "", params)
}

func (as *appService) Get(id string, params Params) (interface{}, error) {
	return as.callMethod(Get, map[string]interface{}{}, id, params)
}

func (as *appService) Create(data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(Create, data, "", params)
}

func (as *appService) Update(id string, data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(Update, data, id, params)
}

func (as *appService) Patch(id string, data map[string]interface{}, params Params) (interface{}, error) {
	return as.callMethod(Patch, data, id, params)
}

func (as *appService) Remove(id string, params Params) (interface{}, error) {
	return as.callMethod(Patch, map[string]interface{}{}, id, params)
}

func (as *appService) HookTree() HooksTree {

	return as.service.HookTree()
}

func (as *appService) callMethod(method RestMethod, data map[string]interface{}, id string, params Params) (interface{}, error) {
	caller := &appServiceCaller{
		success: make(chan interface{}, 0),
		err:     make(chan error, 0),
	}

	as.app.handleServerServiceCall(as.name, Get, caller, data, id, params)
	select {
	case result := <-caller.success:
		return result, nil
	case err := <-caller.err:
		return nil, err
	}
}
