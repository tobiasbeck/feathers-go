package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Discard(fields ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "discard", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch"})
			if err != nil {
				return err
			}
		}

		items, normalized := GetItemsNormalized(ctx)

		for _, item := range items {
			for _, field := range fields {
				delete(item, field)
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return nil
	}
}
