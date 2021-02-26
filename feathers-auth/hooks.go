package feathersAuth

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/fErr"
)

func AuthenticationHook(strategies ...string) feathers.Hook {

	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		service, err := ctx.App.ServiceClass("authentication")
		if err != nil {
			return nil, fErr.Convert(err)
		}
		authService := service.(AuthService)

		return ctx, nil
	}
}
