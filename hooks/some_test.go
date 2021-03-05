package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var someTest = []struct {
	hookChain     []feathers.BoolHook
	expectSuccess bool
	expectError   bool
}{
	/* #1 */ {[]feathers.BoolHook{hookOk()}, true, false},
	/* #2 */ {[]feathers.BoolHook{hookOk(), hookNotOk()}, true, false},
	/* #3 */ {[]feathers.BoolHook{hookNotOk(), hookOk()}, true, false},
	/* #4 */ {[]feathers.BoolHook{hookNotOk(), hookError()}, false, true},
	/* #5 */ {[]feathers.BoolHook{hookOk(), hookError()}, true, false},
	/* #6 */ {[]feathers.BoolHook{hookError(), hookOk()}, false, true},
	/* #7 */ {[]feathers.BoolHook{hookError(), hookNotOk()}, false, true},
}

func TestSome(t *testing.T) {
	for key, data := range someTest {
		context := feathers.HookContext{
			Params: feathers.Params{},
		}

		ok, err := hooks.Some(data.hookChain...)(&context)
		if ok != data.expectSuccess || (err == nil) == data.expectError {
			t.Errorf("Failed #%d: wanted: (ok: %t, wantErr: %t), got: (ok %t, err %v)", key+1, data.expectSuccess, data.expectError, ok, err)
		}
	}
}
