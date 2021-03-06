package feathers_mongo

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MongoKeys(fields ...string) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		if data, ok := ctx.Data.(map[string]interface{}); ok {
			for _, field := range fields {
				value := data[field]
				if strValue, ok := value.(string); ok {
					objId, err := primitive.ObjectIDFromHex(strValue)
					if err != nil {
						data[field] = objId
					}
				}
			}

			return ctx, nil
		}
		return ctx, nil
	}
}
