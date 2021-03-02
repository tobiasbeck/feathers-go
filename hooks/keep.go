package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Keep(keep ...string) feathers.Hook {
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
				for key, _ := range mapData {
					if !contains(keep, key) {
						delete(mapData, key)
					}
				}
			}
		}
		ReplaceItemsNormalized(ctx, items, normalized)
		return ctx, nil
	}
}
