package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

type JoinOperator = func(entity map[string]interface{}, ctx *feathers.Context) error

func Join(joinConfig map[string]JoinOperator) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		data, normalized := GetItemsNormalized(ctx)
		for _, entity := range data {
			for _, operator := range joinConfig {
				err := operator(entity, ctx)
				if err != nil {
					return nil, err
				}

			}
		}
		ReplaceItemsNormalized(ctx, data, normalized)
		return ctx, nil
	}
}
