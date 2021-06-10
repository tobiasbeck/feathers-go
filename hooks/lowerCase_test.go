package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestLowercaseBefore(t *testing.T) {
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "TesT",
			"test2": "TeSt",
		},
	}

	err := hooks.LowerCase("test")(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
	if ctx.Data["test"].(string) != "test" {
		t.Errorf("field not changed correctly. expected 'test', got: %s", ctx.Data["test"].(string))
	}

	if ctx.Data["test2"].(string) != "TeSt" {
		t.Errorf("changed not specified field. expected 'TeSt', got: %s", ctx.Data["test2"].(string))
	}
}
