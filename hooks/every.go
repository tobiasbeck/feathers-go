package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Every(checks ...feathers.BoolHook) feathers.BoolHook {
	return func(ctx *feathers.HookContext) (bool, error) {
		for _, check := range checks {
			ok, err := check(ctx)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	}
}
