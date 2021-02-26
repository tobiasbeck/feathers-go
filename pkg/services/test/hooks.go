package service_test

import (
	"fmt"

	"github.com/tobiasbeck/hackero/pkg/feathers"
)

func testHook(ctx *feathers.HookContext) (*feathers.HookContext, error) {
	fmt.Printf("HOOK TRIGGERECD!")
	ctx.Params.SetField("hook_message", "HELLO FROM A HOOK!")
	return ctx, nil
}

var serviceHooks = feathers.HooksTree{
	Before: feathers.HooksTreeBranch{
		Find:   []feathers.Hook{},
		Get:    []feathers.Hook{},
		Create: []feathers.Hook{},
		Patch:  []feathers.Hook{},
		Update: []feathers.Hook{},
		Remove: []feathers.Hook{},
	},
	After: feathers.HooksTreeBranch{
		Find:   []feathers.Hook{},
		Get:    []feathers.Hook{},
		Create: []feathers.Hook{},
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
