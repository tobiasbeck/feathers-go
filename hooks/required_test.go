package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestRequired(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "HELLO123",
			"test2": "TeSt",
		},
	}

	ctx, err := hooks.Required("test")(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
}

func TestRequired2(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "HELLO123",
			"test2": "TeSt",
		},
	}

	ctx, err := hooks.Required("test", "test3")(ctx)
	if err == nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
}

func TestRequired3(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "HELLO123",
			"test2": "TeSt",
		},
	}

	ctx, err := hooks.Required("test3")(ctx)
	if err == nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
}
