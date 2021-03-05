package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var checkContextTest = []struct {
	hookName       string
	allowedTypes   []feathers.HookType
	allowedMethods []feathers.RestMethod
	ctxType        feathers.HookType
	ctxMethod      feathers.RestMethod
	expectError    bool
}{
	/* #1 */ {"#1", []feathers.HookType{feathers.Before}, []feathers.RestMethod{feathers.Create}, feathers.Before, feathers.Create, false},
}

func TestCheckContext(t *testing.T) {
	for key, data := range checkContextTest {
		context := feathers.HookContext{
			Type:   data.ctxType,
			Method: data.ctxMethod,
		}

		err := hooks.CheckContext(&context, data.hookName, data.allowedTypes, data.allowedMethods)
		if (err != nil) != data.expectError {
			t.Errorf("Failed #%d: wanted: (err: %t), got: ( err %s)", key+1, data.expectError, err)
		}
	}
}
