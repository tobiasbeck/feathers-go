package hooks

import (
	"fmt"
	"time"

	"github.com/tobiasbeck/hackero/pkg/feathers"
)

func MeasureTimeHook() feathers.Hook {
	return func(ctx *feathers.HookContext) (*feathers.HookContext, error) {
		if ctx.Type == feathers.Before {
			ctx.Params.SetField("startTime", time.Now())
			return ctx, nil
		}

		if ctx.Type == feathers.After {
			if startTime, ok := ctx.Params.GetField("startTime"); ok {
				iStartTime := startTime.(time.Time)
				elapsed := time.Since(iStartTime)
				fmt.Printf("EXEC TOOK: %s\n", elapsed)
				return ctx, nil
			}
		}
		return ctx, nil
	}
}
