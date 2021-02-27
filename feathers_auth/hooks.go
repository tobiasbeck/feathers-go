package feathers_auth

import (
	"github.com/tobiasbeck/feathers-go/feathers"
)

func AuthenticationHook(strategies ...string) feathers.Hook {

	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		// service, err := ctx.App.ServiceClass("authentication")
		// if err != nil {
		// 	return nil, feathers_error.Convert(err)
		// }
		// authService := service.(AuthService)

		return ctx, nil
	}
}
