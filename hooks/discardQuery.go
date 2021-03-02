package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func DiscardQuery(fields ...string) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		err := CheckContext(ctx, "discardQuery", []feathers.HookType{"before"}, []feathers.RestMethod{})
		if err != nil {
			return nil, err
		}

		query := ctx.Params.Query

		for _, field := range fields {
			delete(query, field)
		}

		ctx.Params.Query = query
		return ctx, nil
	}
}
