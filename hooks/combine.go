package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Combine(ctx *feathers.HookContext, chain ...feathers.Hook) (*feathers.HookContext, error) {
	var err error
	for _, hook := range chain {
		ctx, err = hook(ctx)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}
