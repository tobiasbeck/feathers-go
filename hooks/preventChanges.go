package hooks

import (
	"strings"

	"github.com/tobiasbeck/feathers-go/feathers"
)

// TODO
func PreventChanges(retError bool, fields ...string) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "discard", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch"})
			if err != nil {
				return nil, err
			}
		}

		items, normalized := GetItemsNormalized(ctx)

		for _, item := range items {
			if mapData, ok := item.(map[string]interface{}); ok {
				for key, value := range mapData {
					if strValue, ok := value.(string); ok {
						mapData[key] = strings.ToLower(strValue)
					}
				}
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return ctx, nil
	}
}
