package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var checkPermissionsTests = []struct {
	userPermissions []string
	hookPermissions []string
	expectError     bool
}{
	/* #1 */ {[]string{"test:create"}, []string{"test:create"}, false},
	/* #2 */ {[]string{"test:patch"}, []string{"test:create"}, true},
	/* #3 */ {[]string{"test:create:any"}, []string{"test:create"}, true},
	/* #4*/ {[]string{"test"}, []string{"test:*"}, false},
	/* #5 */ {[]string{"anye"}, []string{"test:*"}, true},
	/* #6 */ {[]string{"test:patch"}, []string{"test:create"}, true},
	/* #7 */ {[]string{"test"}, []string{"test:get"}, false},
	/* #8 */ {[]string{"*"}, []string{"test:create"}, false},
	/* #9 */ {[]string{"*:get"}, []string{"test:get"}, false},
}

func TestCheckPermissions(t *testing.T) {
	for key, data := range checkPermissionsTests {
		context := feathers.Context{
			Type:   feathers.Before,
			Method: feathers.Get,
			Params: feathers.Params{
				Provider: "socketio",
			},
		}

		context.Params.Set("permissions", data.userPermissions)

		_, err := hooks.CheckPermissions(data.hookPermissions...)(&context)
		if (err == nil) == data.expectError {
			t.Errorf("Failed #%d: wanted: (wantErr: %t), got: (err %v)", key+1, data.expectError, err)
		}
	}
}
