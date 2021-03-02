package hooks

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

/**
  Disallow throws error if user is any of these providers
  If no providers are passed route is completely disabled
*/
func Disallow(providers ...string) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		err := feathers_error.NewMethodNotAllowed("Provider "+ctx.Params.Provider+" can not call "+ctx.Method.String()+". (disallow)", nil)
		if len(providers) == 0 {
			return nil, err
		}

		for _, provider := range providers {
			if provider == "server" && ctx.Params.Provider == "" || provider == "external" && ctx.Params.Provider != "" || provider == ctx.Params.Provider {
				return nil, err
			}
		}

		return ctx, nil
	}
}
