package feathers_mongo

import (
	"github.com/go-playground/validator"
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	*feathers.BaseService
	*feathers.ModelService
	app            *feathers.App
	CollectionName string
	validator      *validator.Validate
}

// Service routes

func (f *Service) Find(params feathers.HookParams) (interface{}, error) {
	if collection, ok := f.getCollection(); ok {

		result, err := collection.Find(params.CallContext, params.Query)
		if err != nil {
			return nil, err
		}

		var returnData []map[string]interface{}
		err = result.All(params.CallContext, &returnData)
		if err != nil {
			return nil, err
		}

		return returnData, err
	}
	return nil, notReady()
}
func (f *Service) Get(id string, params feathers.HookParams) (interface{}, error) {
	return nil, feathers_error.NewNotImplemented("Function is not implemented", nil)
}

func (f *Service) Create(data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	model, err := f.MapToModel(data)
	if err != nil {
		return nil, err
	}

	err = f.ValidateModel(model)
	if err != nil {
		return nil, err
	}

	if collection, ok := f.getCollection(); ok {
		result, err := collection.InsertOne(params.CallContext, model)
		if err != nil {
			return nil, err
		}
		modelMap, err := f.StructToMap(model)
		if err != nil {
			return nil, err
		}
		modelMap["_id"] = result.InsertedID
		// findResult := collection.FindOne(params.CallContext, bson.D{{"_id", result.InsertedID}})
		// var document map[string]interface{}
		// findResult.Decode(&document)
		return modelMap, nil
	}
	return nil, notReady()
}

func (f *Service) Update(id string, data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	return nil, feathers_error.NewNotImplemented("Function is not implemented", nil)
}

func (f *Service) Patch(id string, data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	return nil, feathers_error.NewNotImplemented("Function is not implemented", nil)
}

func (f *Service) Remove(id string, params feathers.HookParams) (interface{}, error) {
	return nil, feathers_error.NewNotImplemented("Function is not implemented", nil)
}

func notReady() error {
	return feathers_error.NewGeneralError("Service not ready", nil)
}

func (f *Service) getCollection() (*mongo.Collection, bool) {
	if db, ok := f.getMongoDb(); ok {
		return db.Collection(f.CollectionName), true
	}
	return nil, false
}

func (f *Service) getMongoDb() (*mongo.Database, bool) {
	if client, ok := f.app.GetConfig("mongoDb"); ok {
		return client.(*mongo.Database), true
	}
	return nil, false
}

func NewService(collection string, model feathers.ModelFactory, app *feathers.App) *Service {
	return &Service{
		BaseService:    &feathers.BaseService{},
		ModelService:   feathers.NewModelService(model),
		CollectionName: collection,

		app: app,
	}
}
