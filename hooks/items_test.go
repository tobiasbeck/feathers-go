package hooks_test

import (
	"reflect"
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestGetItems(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
		Data:   data,
	}

	items := hooks.GetItems(ctx)
	if !reflect.DeepEqual(items, data) {
		t.Errorf("Did not return data correctly")
		return
	}
}

func TestGetItemsAfter(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.After,
		Method: "create",
		Result: data,
	}

	items := hooks.GetItems(ctx)
	if !reflect.DeepEqual(items, data) {
		t.Errorf("Did not return data correctly")
		return
	}
}

func TestReplaceItems(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
	}

	hooks.ReplaceItems(ctx, data)
	if !reflect.DeepEqual(ctx.Data, data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestReplaceItemsAfter(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.After,
		Method: "create",
	}

	hooks.ReplaceItems(ctx, data)
	if !reflect.DeepEqual(ctx.Result, data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestGetItemsNormalized(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
		Data:   data,
	}

	items, normalized := hooks.GetItemsNormalized(ctx)

	if normalized != true {
		t.Errorf("Did not normalize data")
	}

	if !reflect.DeepEqual(items[0], data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestGetItemsNormalizedAfter(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}
	ctx := &feathers.Context{
		Type:   feathers.After,
		Method: "create",
		Result: data,
	}

	items, normalized := hooks.GetItemsNormalized(ctx)

	if normalized != true {
		t.Errorf("Did not normalize data")
	}

	if !reflect.DeepEqual(items[0], data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestReplaceItemsNormalized(t *testing.T) {
	data := []map[string]interface{}{{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}}
	ctx := &feathers.Context{
		Type:   feathers.Before,
		Method: "create",
	}

	hooks.ReplaceItemsNormalized(ctx, data, true)

	if !reflect.DeepEqual(ctx.Data, data[0]) {
		t.Errorf("Did not return data correctly")
	}
}

func TestReplaceItemsNormalizedAfter(t *testing.T) {
	data := []map[string]interface{}{{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}}
	ctx := &feathers.Context{
		Type:   feathers.After,
		Method: "create",
	}

	hooks.ReplaceItemsNormalized(ctx, data, true)

	if !reflect.DeepEqual(ctx.Result, data[0]) {
		t.Errorf("Did not return data correctly")
	}
}
