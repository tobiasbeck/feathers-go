package hooks

import (
	"time"

	"github.com/tobiasbeck/feathers-go/feathers"
)

// SetNow sets fields to Now value (Time.Now()).
func SetNow(fields ...string) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "discard", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch"})
			if err != nil {
				return nil, err
			}
		}

		items, normalized := GetItemsNormalized(ctx)

		for _, item := range items {
			for _, field := range fields {
				item[field] = time.Now()
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return ctx, nil
	}
}
