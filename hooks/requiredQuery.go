package hooks

import (
	"fmt"
	"reflect"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

// Required checks if all passed fields are set
func RequiredQuery(fields ...string) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "requiredQuery", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"find", "create", "update", "patch", "remove"})
			if err != nil {
				return nil, err
			}
		}

		item := ctx.Params.Query

		for _, field := range fields {
			err := feathers_error.NewBadRequest(fmt.Sprintf("Field %s does not exist. (requiredQuery)", field), nil)
			if val, ok := item[field]; ok {
				r := reflect.ValueOf(val)
				if r.IsZero() {
					return nil, err
				}

			} else {
				return nil, err
			}
		}
		return ctx, nil
	}
}