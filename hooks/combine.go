package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func Combine(ctx *feathers.HookContext, chain ...feathers.Hook) (*feathers.HookContext, error) {
	return ctx.App.HandleHookChain(chain, ctx)
}
