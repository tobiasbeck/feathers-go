package hooks_test

import (
	"reflect"
	"testing"

	"github.com/tobiasbeck/feathers-go/hooks"
)

func TestNormalizeSlice(t *testing.T) {
	data := map[string]interface{}{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}

	items, normalized := hooks.NormalizeSlice(data)

	if normalized != true {
		t.Errorf("Did not normalize data")
	}

	if !reflect.DeepEqual(items[0], data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestNormalizeSlice2(t *testing.T) {
	data := []map[string]interface{}{{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}}

	items, normalized := hooks.NormalizeSlice(data)

	if normalized != false {
		t.Errorf("Did normalize data eventhough not required")
	}

	if !reflect.DeepEqual(items, data) {
		t.Errorf("Did not return data correctly")
	}
}

func TestUnnormalizeSlice(t *testing.T) {
	data := []map[string]interface{}{{
		"test":  "TesT",
		"test2": "TeSt",
		"test3": "tEsT",
	}}

	items := hooks.UnormalizeSlice(data, true)

	if !reflect.DeepEqual(items, data[0]) {
		t.Errorf("Did not return data correctly")
	}
}
