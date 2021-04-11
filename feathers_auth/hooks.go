package feathers_auth

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
	"golang.org/x/crypto/bcrypt"
)

func AuthenticationHook(strategies ...string) feathers.Hook {

	return func(ctx *feathers.Context) (*feathers.Context, error) {
		if ctx.Type != feathers.Before {
			return nil, feathers_error.NewNotAuthenticated("The authenticate hook must be used as a before hook", nil)
		}

		if ctx.Params.Provider == "" || (ctx.Params.Connection != nil && ctx.Params.Connection.IsAuthenticated()) {
			return ctx, nil
		}

		// TODO: implement authentication which is not socket based
		// service, err := ctx.App.ServiceClass("authentication")
		// if err != nil {
		// 	return nil, feathers_error.Convert(err)
		// }
		// authService := service.(AuthService)
		// authData := map[string]interface{}{
		// 	strategy: "jwt",

		// }

		// authService.Create(map[string]intewr)

		return nil, feathers_error.NewNotAuthenticated("Not authenticated", nil)
	}
}

// HashPassword is a hool which hashes the password in the given field using bcrypt. If the field is not set or cannot be converted a error is retunred
func HashPassword(field string) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		if password, ok := ctx.Data[field]; ok {
			passwordString, ok := password.(string)
			if ok == false {
				return nil, feathers_error.NewGeneralError("password field incorrect", nil)
			}

			encrypted, err := bcrypt.GenerateFromPassword([]byte(passwordString), 15)
			if err != nil {
				return nil, feathers_error.Convert(err)
			}
			ctx.Data[field] = string(encrypted)
			return ctx, nil
		}
		return nil, feathers_error.NewBadRequest("password not found", nil)
	}
}
