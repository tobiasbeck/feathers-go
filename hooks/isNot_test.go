package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var isNotTest = []struct {
	hook          feathers.BoolHook
	expectSuccess bool
	expectError   bool
}{
	/* #1 */ {hookOk(), false, false},
	/* #2 */ {hookNotOk(), true, false},
	/* #3 */ {hookError(), false, true},
}

func TestIsNot(t *testing.T) {
	for key, data := range isNotTest {
		context := feathers.HookContext{
			Params: feathers.Params{},
		}

		ok, err := hooks.IsNot(data.hook)(&context)
		if ok != data.expectSuccess || (err == nil) == data.expectError {
			t.Errorf("Failed #%d: wanted: (ok: %t, wantErr: %t), got: (ok %t, err %v)", key+1, data.expectSuccess, data.expectError, ok, err)
		}
	}
}
