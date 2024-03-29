package mongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func normalizeArray(data interface{}) interface{} {
	if reflect.TypeOf(data).Kind() == reflect.Slice {
		return data
	}
	return []interface{}{data}
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
	objectIdFields []string
}

// Service routes

func (f *Service) Find(ctx context.Context, params feathers.Params) (interface{}, error) {
	if collection, ok := f.Collection(); ok {
		filters, findOpts, err := f.prepareFilter("", params.Query)
		if err != nil {
			return nil, err
		}

		queryOptions := options.Find()

		if limit, ok := findOpts["$limit"]; ok {
			queryOptions.SetLimit(int64(limit.(int)))
		}
		// fmt.Printf("QUERY: %#v\n\n", filters)
		result, err := collection.Find(ctx, filters, queryOptions)
		if err != nil {
			return nil, err
		}

		var returnData []map[string]interface{}
		err = result.All(ctx, &returnData)
		if err != nil {
			return nil, err
		}

		if returnData == nil {
			returnData = []map[string]interface{}{}
		}
		return normalizeArray(returnData), err
	}
	return nil, notReady()
}
func (f *Service) Get(ctx context.Context, id string, params feathers.Params) (interface{}, error) {
	if collection, ok := f.Collection(); ok {

		query, _, err := f.prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}

		queryOptions := options.Find()
		queryOptions.SetLimit(int64(1))

		result, err := collection.Find(ctx, query, queryOptions)
		if err != nil {
			return nil, err
		}

		var returnData []map[string]interface{}
		err = result.All(ctx, &returnData)
		if err != nil {
			return nil, err
		}
		if len(returnData) <= 0 {
			return nil, httperrors.NewNotFound(fmt.Sprintf("Entity with id %s not found", id), nil)
		}
		// fmt.Printf("\n\nRETURNDATA: %#v\n\n", returnData)
		return returnData[0], err
	}
	return nil, notReady()
}

func (f *Service) Create(ctx context.Context, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	model, err := f.MapToModel(data)
	if err != nil {
		return nil, errors.Wrap(err, "Model Parsing")
	}

	err = f.ValidateModel(model)
	if err != nil {
		return nil, errors.Wrap(err, "Model Validaton")
	}

	if timestampable, ok := model.(Timestampable); ok {
		timestampable.SetCreatedAt()
		timestampable.SetUpdatedAt()
	}
	if idDoc, ok := model.(IdDocument); ok {
		if idDoc.IDIsZero() {
			idDoc.GenerateID()
		}
	}
	if collection, ok := f.Collection(); ok {
		result, err := collection.InsertOne(ctx, model)
		if err != nil {
			return nil, err
		}
		modelMap, err := f.StructToMap(model)
		if err != nil {
			return nil, err
		}
		modelMap["_id"] = result.InsertedID
		params.Set("mongo_result", result)
		// findResult := collection.FindOne(params.CallContext, bson.D{{"_id", result.InsertedID}})
		// var document map[string]interface{}
		// findResult.Decode(&document)
		return modelMap, nil
	}
	return nil, notReady()
}

func (f *Service) Update(ctx context.Context, id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	model, err := f.MapAndValidate(data)
	if err != nil {
		return nil, err
	}

	if timestampable, ok := model.(Timestampable); ok {
		timestampable.SetUpdatedAt()
	}

	if collection, ok := f.Collection(); ok {
		query, _, err := f.prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}

		result, err := collection.ReplaceOne(ctx, query, model)
		modelMap, err := f.StructToMap(model)
		if err != nil {
			return nil, err
		}
		modelMap["_id"] = result.UpsertedID
		params.Set("mongo_result", result)
		// findResult := collection.FindOne(params.CallContext, bson.D{{"_id", result.InsertedID}})
		// var document map[string]interface{}
		// findResult.Decode(&document)
		return modelMap, nil
	}
	return nil, notReady()
}

func (f *Service) Patch(ctx context.Context, id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {

	if collection, ok := f.Collection(); ok {
		query, _, err := f.prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}
		data["updatedAt"] = time.Now()
		replacement := remapModifiers(data)
		// fmt.Printf("replacement: %#v, data: %#v\n", replacement, data)

		opts := options.Update()
		if params.Has("mongodb.upsert") {
			opts.SetUpsert(true)
		}

		result, err := collection.UpdateOne(ctx, query, replacement, opts)
		if err != nil {
			return nil, errors.Wrap(err, "Update Error")
		}
		if result.MatchedCount == 0 && result.UpsertedCount == 0 {
			return nil, httperrors.NewNotFound(fmt.Sprintf("Entity with id %s not found", id), nil)
		}
		params.Set("mongo_result", result)
		findResult := collection.FindOne(ctx, query)
		if findResult.Err() != nil {
			return nil, err
		}
		var document map[string]interface{}
		err = findResult.Decode(&document)
		if err != nil {
			return nil, errors.Wrap(err, "Decode Error")
		}
		return document, nil
	}
	return nil, notReady()
}

func (f *Service) Remove(ctx context.Context, id string, params feathers.Params) (interface{}, error) {
	if collection, ok := f.Collection(); ok {
		query, _, err := f.prepareFilter(id, params.Query)
		if err != nil {
			return nil, err
		}

		findResult := collection.FindOne(ctx, query)
		var document map[string]interface{}
		err = findResult.Decode(&document)
		if err != nil {
			return nil, err
		}

		deleteResult, err := collection.DeleteOne(ctx, query)
		if err != nil {
			return nil, err
		}

		if deleteResult.DeletedCount != 1 {
			return nil, httperrors.NewNotFound("Could not delete entity")
		}
		params.Set("mongo_result", deleteResult)

		return document, nil

	}
	return nil, notReady()
}

func notReady() error {
	return httperrors.NewGeneralError("Service not ready", nil)
}

func (f *Service) Collection() (*mongo.Collection, bool) {
	if db, ok := f.mongoDb(); ok {
		return db.Collection(f.CollectionName), true
	}
	return nil, false
}

func (f *Service) mongoDb() (*mongo.Database, bool) {
	if client, ok := f.app.Config("mongoDb"); ok {
		return client.(*mongo.Database), true
	}
	return nil, false
}

var reservedFilters = []string{"$limit", "$sort", "$select", "$skip"}

func (f *Service) prepareFilter(id string, filter map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {
	feathersFilter := map[string]interface{}{}
	var err error
	if id != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(id)
		if err != nil {
			return filter, feathersFilter, err
		}
	}

	if len(f.objectIdFields) > 0 {
		filter = replaceMongoKeys(filter, f.objectIdFields)
	}

	for filterKey, filerValue := range filter {
		if !strings.HasPrefix(filterKey, "$") {
			continue
		}
		if contains(reservedFilters, filterKey) {
			delete(filter, filterKey)
			feathersFilter["filterKey"] = filerValue
		}
	}
	return filter, feathersFilter, nil
}

func replaceMongoKeys(filter map[string]interface{}, keys []string) map[string]interface{} {
	for key, value := range filter {
		switch v := value.(type) {
		case map[string]interface{}:
			if strings.HasPrefix(key, "$") {
				filter[key] = replaceMongoKeys(v, keys)
			}
		case string:
			if contains(keys, key) {
				objectId, err := primitive.ObjectIDFromHex(v)
				if err != nil {
					continue
				}
				// fmt.Printf("REPLACED %s with %v (err: %s)", key, objectId, err)
				filter[key] = objectId
			}
		}
	}
	return filter
}

func filterObjectIdFields(data map[string]interface{}, keys []string) map[string]interface{} {
	for _, field := range keys {
		value, isSet := data[field]
		if !isSet {
			continue
		}
		switch v := value.(type) {
		case primitive.ObjectID:
			if v == primitive.NilObjectID {
				delete(data, field)
			}
		}
	}
	return data
}

// NewService creates a new mongo service struct
func NewService(collection string, model feathers.ModelFactory, app *feathers.App) *Service {
	service := &Service{
		BaseService:    &feathers.BaseService{},
		ModelService:   feathers.NewModelService(model),
		CollectionName: collection,
		objectIdFields: getModelObjectIdFields(model()),
		app:            app,
	}
	return service
}

func getFirstTagField(tag string) string {
	if strings.Contains(tag, ",") {
		return strings.Split(tag, ",")[0]
	}
	return tag
}

func getModelObjectIdFields(model interface{}) []string {
	fields := []string{}
	e := reflect.ValueOf(model).Elem()
	objectIdType := reflect.TypeOf(primitive.ObjectID{})
	for i := 0; i < e.NumField(); i++ {
		varType := e.Type().Field(i).Type
		if varType.AssignableTo(objectIdType) {
			varName := getFirstTagField(e.Type().Field(i).Tag.Get("mapstructure"))
			// fmt.Printf("\n\n\nvarName: %s\n\n\n", varName)
			if varName == "" {
				varName = strings.ToLower(e.Type().Field(i).Name)
			}
			fields = append(fields, varName)
		}
	}
	return fields
}
