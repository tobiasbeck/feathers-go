package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

type AlterItemHandler = func(item interface{}, ctx *feathers.Context) (interface{}, error)

/*
AlterItems alters the items (either data or result)
different to the original feathers-hook returning nil will remove the item from items
Returning error in handler will immediatly cancel execution and return error
*/
func AlterItems(handler AlterItemHandler) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		itemsInt, normalized := GetItemsNormalized(ctx)
		items := itemsInt.([]map[string]interface{})
		normalizedItems := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			data, err := handler(item, ctx)
			if err != nil {
				return nil, err
			}
			if data != nil {
				normalizedItems = append(normalizedItems, data.(map[string]interface{}))
			}
		}
		ReplaceItemsNormalized(ctx, normalizedItems, normalized)
		return ctx, nil
	}
}
