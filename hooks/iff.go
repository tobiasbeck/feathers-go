package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Iff(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return ctx, nil
		}

		for _, hook := range trueHooks {
			ctx, err = hook(ctx)
			if err != nil {
				return nil, err
			}
		}
		return ctx, nil
	}
}

func IffElse(pred feathers.BoolHook, trueHooks []feathers.Hook, falseHooks []feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			for _, hook := range falseHooks {
				ctx, err = hook(ctx)
				if err != nil {
					return nil, err
				}
			}
			return ctx, nil
		}

		for _, hook := range trueHooks {
			ctx, err = hook(ctx)
			if err != nil {
				return nil, err
			}
		}
		return ctx, nil
	}
}

func IffNot(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) (*feathers.Context, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			return ctx, nil
		}

		for _, hook := range trueHooks {
			ctx, err = hook(ctx)
			if err != nil {
				return nil, err
			}
		}
		return ctx, nil
	}
}
