package hooks

import (
	"errors"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

func CheckPermissions(requiredPermissions ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		if ctx.Type != feathers.Before {
			return errors.New("The feathers-permissions hook should only be used as a 'before' hook")
		}
		hookPermissions, ok := ctx.Params.Lookup("permissions")
		if !ok {
			if ok, _ := IsProvider("external")(ctx); ok {
				return feathers_error.NewForbidden("You do not have the correct permissions (invalid permission entity).", nil)
			}
		}
		var currentPermissions []string
		switch p := hookPermissions.(type) {
		case string:
			currentPermissions = []string{p}
		case []string:
			currentPermissions = p
		default:
			return feathers_error.NewGeneralError("You do ont have the correct permissions (permission datatype mismatch)", nil)
		}

		requiredPermissionsWildcards := append([]string{}, "*", "*:"+ctx.Method.String())
		requiredPermissionsWildcards = append(requiredPermissionsWildcards, requiredPermissions...)

		for _, permission := range currentPermissions {
			permissionWildcards := []string{permission, permission + ":*", permission + ":" + ctx.Method.String()}
			for _, p := range permissionWildcards {
				if contains(requiredPermissionsWildcards, p) {
					return nil
				}
			}
		}
		return feathers_error.NewForbidden("You do not have the correct permissions.", nil)
	}
}
