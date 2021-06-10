package hooks

import (
	"fmt"
	"reflect"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

// Required checks if all passed fields are set
func Required(fields ...string) feathers.Hook {
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
				err := feathers_error.NewBadRequest(fmt.Sprintf("Field %s does not exist. (preventChanges)", field), nil)
				if val, ok := item[field]; ok {
					r := reflect.ValueOf(val)
					if r.IsZero() {
						return err
					}

				} else {
					return err
				}
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return nil
	}
}
