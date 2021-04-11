package hooks_test

import (
	"testing"
	"time"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestSetNow(t *testing.T) {
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "HELLO123",
			"test2": "TeSt",
		},
	}

	ctx, err := hooks.SetNow("test", "test3")(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
	if _, ok := ctx.Data["test"].(time.Time); !ok {
		t.Errorf("field not changed correctly. expected value of time.Time, got: %t", ctx.Data["test"])
	}

	if _, ok := ctx.Data["test3"].(time.Time); !ok {
		t.Errorf("field not changed correctly. expected value of time.Time, got: %t", ctx.Data["test"])
	}
}
