package hooks_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestAlterItemsBefore(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "TesT",
			"test2": "TeSt",
			"test3": "TaTe",
		},
	}

	ctx, err := hooks.AlterItems(func(item interface{}, ctx *feathers.HookContext) (interface{}, error) {
		if itemMap, ok := item.(map[string]interface{}); ok {
			delete(itemMap, "test")
			itemMap["test2"] = "hello"
			return item, nil
		}
		return nil, errors.New("could not transform item to map[string]interface{}")
	})(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}
	if _, ok := ctx.Data.(map[string]interface{})["test"]; ok {
		t.Errorf("field not removed correctly. expected field to be missing, got: %s", ctx.Data.(map[string]interface{})["test"].(string))
	}

	if ctx.Data.(map[string]interface{})["test2"].(string) != "hello" {
		t.Errorf("field not changed correclty. expected 'hello', got: %s", ctx.Data.(map[string]interface{})["test2"].(string))
	}

	if ctx.Data.(map[string]interface{})["test3"].(string) != "TaTe" {
		t.Errorf("changed not specified field. expected 'TeSt', got: %s", ctx.Data.(map[string]interface{})["test3"].(string))
	}
}

func TestAlterItemsError(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data:   map[string]interface{}{},
	}

	ctx, err := hooks.AlterItems(func(item interface{}, ctx *feathers.HookContext) (interface{}, error) {
		return nil, errors.New("could not transform item to map[string]interface{}")
	})(ctx)
	if err == nil {
		t.Errorf("Hook did not return required error")
		return
	}
}

func TestAlterItemsRemove(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: map[string]interface{}{
			"test":  "TesT",
			"test2": "TeSt",
			"test3": "TaTe",
		},
	}

	ctx, err := hooks.AlterItems(func(item interface{}, ctx *feathers.HookContext) (interface{}, error) {
		return nil, nil
	})(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}

	if ctx.Data != nil {
		t.Errorf("entity not removed correctly")
	}
}

func TestAlterItemsRemoveArray(t *testing.T) {
	ctx := &feathers.HookContext{
		Type:   feathers.Before,
		Method: "create",
		Data: []map[string]interface{}{{
			"test":  "TesT",
			"test2": "TeSt",
			"test3": "TaTe",
		},
			{
				"test":  "remove",
				"test2": "TeSt",
				"test3": "TaTe",
			},
		},
	}

	ctx, err := hooks.AlterItems(func(item interface{}, ctx *feathers.HookContext) (interface{}, error) {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["test"].(string) == "remove" {
				return nil, nil
			}
			return item, nil
		}
		return nil, fmt.Errorf("could not transform item to map[string]interface{}, %T", item)
	})(ctx)
	if err != nil {
		t.Errorf("Hook returned unexpected error: %s", err)
		return
	}

	if len(ctx.Data.([]map[string]interface{})) != 1 {
		t.Errorf("entity not removed correctly")
	}
}

// func TestLowercaseAfter(t *testing.T) {
// 	ctx := &feathers.HookContext{
// 		Type:   feathers.Before,
// 		Method: "after",
// 		Result: map[string]interface{}{
// 			"test":  "TesT",
// 			"test2": "TeSt",
// 		},
// 	}

// 	ctx, err := hooks.LowerCase("test")(ctx)
// 	if err != nil {
// 		t.Errorf("Hook returned unexpected error: %s", err)
// 		return
// 	}
// 	if ctx.Result.(map[string]interface{})["test"].(string) != "test" {
// 		t.Errorf("field not changed correctly. expected 'test', got: %s", ctx.Result.(map[string]interface{})["test"].(string))
// 	}

// 	if ctx.Result.(map[string]interface{})["test2"].(string) != "TeSt" {
// 		t.Errorf("changed not specified field. expected 'TeSt', got: %s", ctx.Result.(map[string]interface{})["test"].(string))
// 	}
// }
