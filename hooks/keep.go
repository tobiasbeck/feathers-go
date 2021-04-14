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
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "keep", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch"})
			if err != nil {
				return nil, err
			}
		}

		items, normalized := GetItemsNormalized(ctx)

		for _, item := range items.([]map[string]interface{}) {
			for key, _ := range item {
				if !contains(keep, key) {
					delete(item, key)
				}
			}
		}
		ReplaceItemsNormalized(ctx, items, normalized)
		return ctx, nil
	}
}
