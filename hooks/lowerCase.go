package hooks

import (
	"strings"

	"github.com/tobiasbeck/feathers-go/feathers"
)

func LowerCase(fields ...string) feathers.Hook {
	return func(ctx *feathers.Context) error {
		if ctx.Type == feathers.Before {
			err := CheckContext(ctx, "lowerCase", []feathers.HookType{"before", "after"}, []feathers.RestMethod{"create", "update", "patch"})
			if err != nil {
				return err
			}
		}

		items, normalized := GetItemsNormalized(ctx)
		for _, item := range items {
			for _, field := range fields {
				value := item[field]
				if strValue, ok := value.(string); ok {
					item[field] = strings.ToLower(strValue)
				}
			}
		}

		ReplaceItemsNormalized(ctx, items, normalized)
		return nil
	}
}
