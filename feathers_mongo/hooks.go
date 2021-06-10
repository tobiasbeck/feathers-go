package feathers_mongo

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MongoKeys(fields ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		data := ctx.Data
		for _, field := range fields {
			value := data[field]
			if strValue, ok := value.(string); ok {
				objId, err := primitive.ObjectIDFromHex(strValue)
				if err != nil {
					data[field] = objId
				}
			}
		}

		return nil
	}
}
