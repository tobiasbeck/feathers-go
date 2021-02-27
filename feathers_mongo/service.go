package feathers_mongo

import (
	"fmt"

	"github.com/go-playground/validator"
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func prepareFilter(id string, filter map[string]interface{}) (map[string]interface{}, error) {
	var err error
	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return filter, err
	}
	return filter, nil
}

type m = map[string]interface{}

func mergeKeys(left, right m) m {
	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			//then we don't want to replace it - recurse
			left[key] = mergeKeys(leftVal.(m), rightVal.(m))
		} else {
			// key not in left so we can just shove it in
			left[key] = rightVal
		}
	}
	return left
}

func remapModifiers(filter map[string]interface{}) map[string]interface{} {
	set := make(map[string]interface{})
	remapped := make(map[string]interface{})
	for key, value := range filter {
		if key[0] != '$' {
			set[key] = value
			continue
		}
		if key == "$set" {
			set = mergeKeys(set, value.(map[string]interface{}))
			continue
		}
		remapped[key] = value
	}

	if len(set) > 0 {
		remapped["$set"] = set
	}
	return remapped
}

// Service for mongodb which offers model validation. Use `NewService` for new instance. (another service is supposed to extend from this)
type Service struct {
	*feathers.BaseService
	*feathers.ModelService
	app            *feathers.App
	CollectionName string
	validator      *validator.Validate
}

// Service routes

func (f *Service) Find(params feathers.Params) (interface{}, error) {
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
func (f *Service) Get(id string, params feathers.Params) (interface{}, error) {
	if collection, ok := f.getCollection(); ok {

		query, err := prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}
		result, err := collection.Find(params.CallContext, query)
		if err != nil {
			return nil, err
		}

		var returnData []map[string]interface{}
		err = result.All(params.CallContext, &returnData)
		if err != nil {
			return nil, err
		}

		return returnData[0], err
	}
	return nil, notReady()
}

func (f *Service) Create(data map[string]interface{}, params feathers.Params) (interface{}, error) {
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

func (f *Service) Update(id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	fmt.Printf("HELLO\n")
	model, err := f.MapAndValidate(data)
	if err != nil {
		return nil, err
	}

	if collection, ok := f.getCollection(); ok {
		query, err := prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}

		result, err := collection.ReplaceOne(params.CallContext, query, model)
		modelMap, err := f.StructToMap(model)
		if err != nil {
			return nil, err
		}
		modelMap["_id"] = result.UpsertedID
		// findResult := collection.FindOne(params.CallContext, bson.D{{"_id", result.InsertedID}})
		// var document map[string]interface{}
		// findResult.Decode(&document)
		return modelMap, nil
	}
	return nil, notReady()
}

func (f *Service) Patch(id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {

	if collection, ok := f.getCollection(); ok {
		query, err := prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}
		replacement := remapModifiers(data)
		fmt.Printf("replacement: %#v\n", replacement)

		result, err := collection.UpdateOne(params.CallContext, query, replacement)
		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			return nil, nil
		}
		findResult := collection.FindOne(params.CallContext, query)
		var document map[string]interface{}
		findResult.Decode(&document)
		return document, nil
	}
	return nil, notReady()
}

func (f *Service) Remove(id string, params feathers.Params) (interface{}, error) {
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
	if client, ok := f.app.Config("mongoDb"); ok {
		return client.(*mongo.Database), true
	}
	return nil, false
}

// NewService creates a new mongo service struct
func NewService(collection string, model feathers.ModelFactory, app *feathers.App) *Service {
	return &Service{
		BaseService:    &feathers.BaseService{},
		ModelService:   feathers.NewModelService(model),
		CollectionName: collection,

		app: app,
	}
}
