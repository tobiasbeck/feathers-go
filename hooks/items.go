package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func GetItems(ctx *feathers.Context) interface{} {
	if ctx.Type == feathers.Before {
		return ctx.Data
	} else {
		return ctx.Result
	}
}

func ReplaceItems(ctx *feathers.Context, data interface{}) {
	if ctx.Type == feathers.Before {
		ctx.Data = data.(map[string]interface{})
	} else {
		ctx.Result = data
	}
}

func GetItemsNormalized(ctx *feathers.Context) (interface{}, bool) {
	if ctx.Type == feathers.Before {
		return NormalizeSlice(ctx.Data)
	} else {
		return NormalizeSlice(ctx.Result)
	}
}

func ReplaceItemsNormalized(ctx *feathers.Context, data interface{}, normalized bool) {
	normData := UnormalizeSlice(data, normalized)
	if ctx.Type == feathers.Before {
		ctx.Data = normData.(map[string]interface{})
	} else {
		ctx.Result = normData
	}
}
