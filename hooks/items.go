package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func GetItems(ctx *feathers.HookContext) interface{} {
	if ctx.Type == feathers.Before {
		return ctx.Data
	} else {
		return ctx.Result
	}
}

func ReplaceItems(ctx *feathers.HookContext, data interface{}) {
	if ctx.Type == feathers.Before {
		ctx.Data = data
	} else {
		ctx.Result = data
	}
}

func GetItemsNormalized(ctx *feathers.HookContext) ([]map[string]interface{}, bool) {
	if ctx.Type == feathers.Before {
		return NormalizeSlice(ctx.Data)
	} else {
		return NormalizeSlice(ctx.Result)
	}
}

func ReplaceItemsNormalized(ctx *feathers.HookContext, data []map[string]interface{}, normalized bool) {
	normData := UnormalizeSlice(data, normalized)
	if ctx.Type == feathers.Before {
		ctx.Data = normData
	} else {
		ctx.Result = normData
	}
}
