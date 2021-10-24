package auth

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
	"golang.org/x/crypto/bcrypt"
)

func AuthenticationHook(strategies ...string) feathers.Hook {

	return func(ctx *feathers.Context) error {
		if ctx.Type != feathers.Before {
			return httperrors.NewNotAuthenticated("The authenticate hook must be used as a before hook", nil)
		}

		if ctx.Params.Provider == "" || (ctx.Params.Connection != nil && ctx.Params.Connection.IsAuthenticated()) {
			return nil
		}

		// TODO: implement authentication which is not socket based
		// service, err := ctx.App.ServiceClass("authentication")
		// if err != nil {
		// 	return nil, httperrors.Convert(err)
		// }
		// authService := service.(AuthService)
		// authData := map[string]interface{}{
		// 	strategy: "jwt",

		// }

		// authService.Create(map[string]intewr)

		return httperrors.NewNotAuthenticated("Not authenticated", nil)
	}
}

// HashPassword is a hool which hashes the password in the given field using bcrypt. If the field is not set or cannot be converted a error is retunred
func HashPassword(field string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		if password, ok := ctx.Data[field]; ok {
			passwordString, ok := password.(string)
			if ok == false {
				return httperrors.NewGeneralError("password field incorrect", nil)
			}

			encrypted, err := bcrypt.GenerateFromPassword([]byte(passwordString), 15)
			if err != nil {
				return httperrors.Convert(err)
			}
			ctx.Data[field] = string(encrypted)
			return nil
		}
		return httperrors.NewBadRequest("password not found", nil)
	}
}
