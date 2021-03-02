package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

//IsProvider BoolHook checks if is provider of specific type
func IsProvider(provider string) feathers.BoolHook {
	return func(ctx *feathers.HookContext) (bool, error) {
		if provider == "server" && ctx.Params.Provider == "" {
			return true, nil
		}
		if provider == "external" && ctx.Params.Provider != "" {
			return true, nil
		}
		return ctx.Params.Provider == provider, nil
	}
}
