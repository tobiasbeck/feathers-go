package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestKeep(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "TesT",
			"test2": "TeSt",
			"test3": "tEsT",
		},
	}

	ctx, err := hooks.Keep("test", "test2")(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
	if _, ok := ctx.Data.(map[string]interface{})["test"].(string); !ok {
		t.Errorf("required field was removed. expected test to be defined, but was not defined")
	}

	if _, ok := ctx.Data.(map[string]interface{})["test2"].(string); !ok {
		t.Errorf("required field was removed. expected test to be defined, but was not defined")
	}

	if _, ok := ctx.Data.(map[string]interface{})["test3"].(string); ok {
		t.Errorf("field to remove still exists. expected test to be removed, but was still defined")
	}
}
