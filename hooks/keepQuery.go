package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func KeepQuery(keep ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "keepquery", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch", "find"})
			if err != nil {
				return err
			}
		}

		item := ctx.Params.Query

		for key, _ := range item {
			if !contains(keep, key) {
				delete(item, key)
			}
		}
		return nil
	}
}
