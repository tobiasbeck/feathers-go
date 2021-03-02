package hooks

import (
	"errors"

	"github.com/tobiasbeck/feathers-go/feathers"
)

/*
CheckContext checks the context for required properties. returns error if requirement not set.
*/
func CheckContext(ctx *feathers.HookContext, hookName string, allowedTypes []feathers.HookType, allowedMethods []feathers.RestMethod) error {
	if len(allowedTypes) > 0 {
		found := false
		for _, allowedType := range allowedTypes {
			if allowedType == ctx.Type {
				found = true
			}
		}
		if found == false {
			return errors.New("You cannot use " + hookName + " as " + ctx.Method.String() + " hook")
		}
	}

	if len(allowedMethods) > 0 {
		found := false
		for _, allowedMethod := range allowedMethods {
			if allowedMethod == ctx.Method {
				found = true
			}
		}
		if found == false {
			return errors.New("You cannot use " + hookName + " in " + ctx.Method.String() + " method")
		}
	}
	return nil
}
