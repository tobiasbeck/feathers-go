package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Iff(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) error {
		ok, err := pred(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		for _, hook := range trueHooks {
			err = hook(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func IffElse(pred feathers.BoolHook, trueHooks []feathers.Hook, falseHooks []feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) error {
		ok, err := pred(ctx)
		if err != nil {
			return err
		}
		if !ok {
			for _, hook := range falseHooks {
				err = hook(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}

		for _, hook := range trueHooks {
			err = hook(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func IffNot(pred feathers.BoolHook, trueHooks ...feathers.Hook) feathers.Hook {
	return func(ctx *feathers.Context) error {
		ok, err := pred(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		for _, hook := range trueHooks {
			err = hook(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
