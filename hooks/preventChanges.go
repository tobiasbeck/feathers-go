package hooks

import (
	"fmt"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

func PreventChanges(retError bool, fields ...string) feathers.Hook {
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
				if _, ok := item[field]; ok {
					if retError {
						return feathers_error.NewBadGateway(fmt.Sprintf("Field %s may not be patched. (preventChanges)", field), nil)
					}
					delete(item, field)

				}
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return nil
	}
}
