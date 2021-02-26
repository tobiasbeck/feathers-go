package feathers

import (
	"strings"

	"github.com/go-playground/validator"
	"github.com/mitchellh/mapstructure"
)

type Service interface {
	Find(params HookParams) (interface{}, error)
	Get(id string, params HookParams) (interface{}, error)

	Create(data map[string]interface{}, params HookParams) (interface{}, error)

	Update(id string, data map[string]interface{}, params HookParams) (interface{}, error)

	Patch(id string, data map[string]interface{}, params HookParams) (interface{}, error)

	Remove(id string, params HookParams) (interface{}, error)

	GetHooks() HooksTree
}

type HooksTreeBranch struct {
	Find   []Hook
	Get    []Hook
	Create []Hook
	Patch  []Hook
	Update []Hook
	Remove []Hook
}

func (b HooksTreeBranch) GetBranch(method CallMethod) []Hook {
	key := strings.Title(method.String())
	// fmt.Printf("checkBranch %#v\n", b)
	if chain, ok := getField(&b, key); ok == true {
		return chain.([]Hook)
	}
	return make([]Hook, 0)
}

type HooksTree struct {
	Before HooksTreeBranch
	After  HooksTreeBranch
	Error  HooksTreeBranch
}

type BaseService struct {
	Hooks HooksTree
}

func (b *BaseService) GetHooks() HooksTree {
	return b.Hooks
}

type ModelFactory = func() interface{}
type ModelService struct {
	Model     ModelFactory
	validator *validator.Validate
}

func (m *ModelService) MapToModel(data map[string]interface{}) (interface{}, error) {
	model := m.Model()
	err := mapstructure.Decode(data, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (m *ModelService) MapToStruct(data map[string]interface{}, target interface{}) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	return nil
}

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

func (m *ModelService) MapAndValidateStruct(data map[string]interface{}, target interface{}) error {
	err := m.MapToStruct(data, target)
	if err != nil {
		return err
	}
	err = m.ValidateModel(target)
	if err != nil {
		return err
	}
	return nil
}

func (m *ModelService) ValidateModel(model interface{}) error {
	err := m.validator.Struct(model)
	return err
}

func NewModelService(model ModelFactory) *ModelService {
	return &ModelService{
		Model:     model,
		validator: validator.New(),
	}
}
