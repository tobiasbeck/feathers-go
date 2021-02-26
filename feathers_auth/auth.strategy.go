package feathers_auth

import "github.com/tobiasbeck/feathers-go/feathers"

type AuthStrategy interface {
	Authenticate(data Model, params feathers.HookParams) (map[string]interface{}, error)
	SetApp(app *feathers.App)
	SetConfiguration(config map[string]interface{})
}

type BaseAuthStrategy struct {
	app    *feathers.App
	config map[string]interface{}
}

func (bas *BaseAuthStrategy) SetApp(app *feathers.App) {
	bas.app = app
}

func (bas *BaseAuthStrategy) SetConfiguration(config map[string]interface{}) {
	bas.config = config
}
