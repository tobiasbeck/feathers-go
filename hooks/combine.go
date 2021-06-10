package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Combine(ctx *feathers.Context, chain ...feathers.Hook) error {
	var err error
	for _, hook := range chain {
		err = hook(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
