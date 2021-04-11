package hooks_test

import (
	"errors"
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

func hookOk() feathers.BoolHook {
	return func(ctx *feathers.Context) (bool, error) {
		return true, nil
	}
}

func hookNotOk() feathers.BoolHook {
	return func(ctx *feathers.Context) (bool, error) {
		return false, nil
	}
}

func hookError() feathers.BoolHook {
	return func(ctx *feathers.Context) (bool, error) {
		return false, errors.New("Any error")
	}
}

var everyTest = []struct {
	hookChain     []feathers.BoolHook
	expectSuccess bool
	expectError   bool
}{
	/* #1 */ {[]feathers.BoolHook{hookOk()}, true, false},
	/* #2 */ {[]feathers.BoolHook{hookOk(), hookNotOk()}, false, false},
	/* #3 */ {[]feathers.BoolHook{hookNotOk(), hookOk()}, false, false},
	/* #4 */ {[]feathers.BoolHook{hookNotOk(), hookError()}, false, false},
	/* #5 */ {[]feathers.BoolHook{hookOk(), hookError()}, false, true},
	/* #6 */ {[]feathers.BoolHook{hookError(), hookOk()}, false, true},
	/* #7 */ {[]feathers.BoolHook{hookError(), hookNotOk()}, false, true},
}

func TestEvery(t *testing.T) {
	for key, data := range everyTest {
		context := feathers.Context{
			Params: feathers.Params{},
		}

		ok, err := hooks.Every(data.hookChain...)(&context)
		if ok != data.expectSuccess || (err == nil) == data.expectError {
			t.Errorf("Failed #%d: wanted: (ok: %t, wantErr: %t), got: (ok %t, err %v)", key+1, data.expectSuccess, data.expectError, ok, err)
		}
	}
}
