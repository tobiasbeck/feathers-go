package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var disallowTest = []struct {
	provider     string
	hookProvider []string
	expectError  bool
}{
	/* #1 */ {"socketio", []string{"socketio"}, true},
	/* #2 */ {"socketio", []string{"http"}, false},
	/* #3 */ {"socketio", []string{"server"}, false},
	/* #4 */ {"socketio", []string{"external"}, true},
	/* #5 */ {"", []string{"server"}, true},
	/* #6 */ {"socketio", []string{}, true},
	/* #7 */ {"", []string{}, true},
	/* #8 */ {"", []string{"socketio"}, false},
}

func TestDisallow(t *testing.T) {
	for key, data := range disallowTest {
		context := feathers.Context{
			Params: feathers.Params{
				Provider: data.provider,
			},
		}

		err := hooks.Disallow(data.hookProvider...)(&context)
		if (err == nil) == data.expectError {
			t.Errorf("Failed #%d: wanted: (wantErr: %t), got: (err %v)", key+1, data.expectError, err)
		}
	}
}
