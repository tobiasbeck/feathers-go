package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func IsNot(check feathers.BoolHook) feathers.BoolHook {
	return func(ctx *feathers.HookContext) (bool, error) {
		ok, err := check(ctx)
		if err != nil {
			return false, err
		}
		return !ok, nil
	}
}
