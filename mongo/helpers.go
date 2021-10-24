package mongo

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func NormalizeObjectId(id interface{}) primitive.ObjectID {
	if oId, ok := id.(primitive.ObjectID); ok {
		return oId
	}
	if sId, ok := id.(string); ok {
		oId, err := primitive.ObjectIDFromHex(sId)
		if err != nil {
			return primitive.NilObjectID
		}
		return oId
	}

	if sId, ok := id.(fmt.Stringer); ok {
		oId, err := primitive.ObjectIDFromHex(sId.String())
		if err != nil {
			return primitive.NilObjectID
		}
		return oId
	}
	return primitive.NilObjectID
}

func objectIdString(id interface{}) string {
	if oId, ok := id.(primitive.ObjectID); ok {
		return oId.Hex()
	}
	if sId, ok := id.(string); ok {
		return sId
	}

	if sId, ok := id.(fmt.Stringer); ok {
		return sId.String()
	}
	return ""
}

func ObjectIDEquals(id interface{}, id2 interface{}) bool {
	idString := objectIdString(id)
	id2String := objectIdString(id2)

	return idString == id2String
}

func MapDecodeMongo() mapstructure.DecodeHookFunc {
	return func(s reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if s == reflect.TypeOf(primitive.ObjectID{}) {
			fObjId := data.(primitive.ObjectID)
			// fmt.Printf("CHECK 1 %s, %s --> %s\n\n", fObjId, s.Name(), t.Name())
			if fObjId == primitive.NilObjectID {
				// rf := reflect.ValueOf(data)
				// fmt.Printf("is Nil %+v\n\n", [0]byte{})
				return nil, nil
			}
			if t.Kind() == reflect.String {
				return fObjId.Hex(), nil
			}
			return data, nil
		}
		// fmt.Println("DECODE MONGO CALL", f.Name(), t.Name())
		if s.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(primitive.ObjectID{}) {
			return data, nil
		}
		result, err := primitive.ObjectIDFromHex(data.(string))
		if err != nil {
			return nil, err
		}
		// fmt.Printf("Add mongoId %s\n\n", result)
		return result, nil
	}
}
