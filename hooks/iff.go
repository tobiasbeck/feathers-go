package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Iff(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return ctx, nil
		}

		return ctx.App.HandleHookChain(trueHooks, ctx)
	}
}

func IffElse(pred feathers.BoolHook, trueHooks []feathers.Hook, falseHooks []feathers.Hook) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return ctx.App.HandleHookChain(falseHooks, ctx)
		}

		return ctx.App.HandleHookChain(trueHooks, ctx)
	}
}

func IffNot(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		ok, err := pred(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			return ctx, nil
		}

		return ctx.App.HandleHookChain(trueHooks, ctx)
	}
}
