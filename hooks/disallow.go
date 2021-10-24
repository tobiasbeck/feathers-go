package hooks

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
)

/**
  Disallow throws error if user is any of these providers
  If no providers are passed route is completely disabled
*/
func Disallow(providers ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		err := httperrors.NewMethodNotAllowed("Provider "+ctx.Params.Provider+" can not call "+ctx.Method.String()+". (disallow)", nil)
		if len(providers) == 0 {
			return err
		}

		for _, provider := range providers {
			if (provider == "server" && ctx.Params.Provider == "") || (provider == "external" && ctx.Params.Provider != "") || provider == ctx.Params.Provider {
				return err
			}
		}

		return nil
	}
}
