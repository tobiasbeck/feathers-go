package hooks_test

import (
	"testing"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/hooks"
)

var providerTest = []struct {
	provider     string
	hookProvider string
	resultOk     bool
	resultErr    error
}{
	/* #1 */ {"", "server", true, nil},
	/* #2 */ {"", "", false, nil},
	/* #3 */ {"", "external", false, nil},
	/* #4 */ {"", "socketio", false, nil},
	/* #5 */ {"socketio", "socketio", true, nil},
	/* #6 */ {"socketio", "external", true, nil},
	/* #7 */ {"socketio", "server", false, nil},
	/* #8 */ {"http", "external", true, nil},
	/* #9 */ {"http", "socketio", false, nil},
}

func TestProvider(t *testing.T) {
	for key, data := range providerTest {
		context := feathers.Context{
			Params: feathers.Params{
				Provider: data.provider,
			},
		}

		ok, err := hooks.IsProvider(data.hookProvider)(&context)
		if ok != data.resultOk || err != data.resultErr {
			t.Errorf("Failed #%d: wanted: (ok: %t, err: %s), got: (ok: %t, err %s)", key+1, data.resultOk, data.resultErr, ok, err)
		}
	}
}
