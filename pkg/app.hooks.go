package app

import (
	"fmt"

	"github.com/tobiasbeck/hackero/app/hooks"
	"github.com/tobiasbeck/hackero/pkg/feathers"
)

func testHook(ctx *feathers.HookContext) (*feathers.HookContext, error) {
	fmt.Printf("HOOK TRIGGERECD!")
	ctx.Params.SetField("hook_message", "HELLO FROM A HOOK2!")
	return ctx, nil
}

var AppHooks = feathers.HooksTree{
	Before: feathers.HooksTreeBranch{
		Find:   []feathers.Hook{},
		Get:    []feathers.Hook{},
		Create: []feathers.Hook{hooks.MeasureTimeHook()},
		Patch:  []feathers.Hook{},
		Update: []feathers.Hook{},
		Remove: []feathers.Hook{},
	},
	After: feathers.HooksTreeBranch{
		Find:   []feathers.Hook{},
		Get:    []feathers.Hook{},
		Create: []feathers.Hook{hooks.MeasureTimeHook()},
		Patch:  []feathers.Hook{},
		Update: []feathers.Hook{},
		Remove: []feathers.Hook{},
	},
	Error: feathers.HooksTreeBranch{
		Find:   []feathers.Hook{},
		Get:    []feathers.Hook{},
		Create: []feathers.Hook{},
		Patch:  []feathers.Hook{},
		Update: []feathers.Hook{},
		Remove: []feathers.Hook{},
	},
}
