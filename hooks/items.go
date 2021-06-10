package hooks

import "github.com/tobiasbeck/feathers-go/feathers"

func normalizeToMapSlice(slice interface{}) []map[string]interface{} {
	switch v := slice.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(v))
		for _, m := range v {
			mm, ok := m.(map[string]interface{})
			if !ok {
				continue
			}
			result = append(result, mm)
		}
		return result
	}
	return []map[string]interface{}{}
}

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

func GetItemsNormalized(ctx *feathers.Context) ([]map[string]interface{}, bool) {
	if ctx.Type == feathers.Before {
		slice, normalized := NormalizeSlice(ctx.Data)
		mapSlice := normalizeToMapSlice(slice)
		return mapSlice, normalized
	} else {
		slice, normalized := NormalizeSlice(ctx.Result)
		mapSlice := normalizeToMapSlice(slice)
		return mapSlice, normalized
	}
}

func ReplaceItemsNormalized(ctx *feathers.Context, data interface{}, normalized bool) {
	normData := UnormalizeSlice(data, normalized)
	if ctx.Type == feathers.Before {
		if normData != nil {
			ctx.Data = normData.(map[string]interface{})
		} else {
			ctx.Data = nil
		}

	} else {
		ctx.Result = normData
	}
}
