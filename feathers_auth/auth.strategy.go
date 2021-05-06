package feathers_auth

import (
	"context"
	"errors"

	defaults "github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/tobiasbeck/feathers-go/feathers"
)

type DefaultAuthConfig struct {
	Entity string `mapstructure:"entity" default:"entity"`
	Secret string `mapstructure:"secret"`
}

type AuthStrategy interface {
	Authenticate(ctx context.Context, data Model, params feathers.Params) (map[string]interface{}, error)
	SetApp(app *feathers.App)
	SetName(name string)
	SetConfiguration(config map[string]interface{})
}

type BaseAuthStrategy struct {
	app    *feathers.App
	config map[string]interface{}
	name   string
}

func (bas *BaseAuthStrategy) SetApp(app *feathers.App) {
	bas.app = app
}

func (bas *BaseAuthStrategy) SetName(name string) {
	bas.name = name
}

func (bas *BaseAuthStrategy) SetConfiguration(config map[string]interface{}) {
	bas.config = config
}

func (bas *BaseAuthStrategy) Config(key string) (interface{}, bool) {
	if config, ok := bas.config[key]; ok {
		return config, true
	}
	return nil, false
}

func (bas *BaseAuthStrategy) StrategyConfig(config interface{}) error {
	if configMap, ok := bas.Config(bas.name); ok {
		mapstructure.Decode(configMap, config)
		defaults.SetDefaults(config)
		return nil
	}
	return errors.New("Key " + bas.name + " is not set")
}

func (bas *BaseAuthStrategy) EntityService() (feathers.Service, bool) {
	if serviceName, ok := bas.Config("service"); ok {
		service := bas.app.Service(serviceName.(string))
		if service == nil {
			return nil, false
		}
		return service, true
	}
	return nil, false
}

func (bas *BaseAuthStrategy) DefaultConfig() DefaultAuthConfig {
	config := DefaultAuthConfig{}
	mapstructure.Decode(bas.config, &config)
	defaults.SetDefaults(&config)
	return config
}

func NewBaseAuthStrategy(name string, app *feathers.App) *BaseAuthStrategy {
	return &BaseAuthStrategy{
		name:   name,
		app:    app,
		config: map[string]interface{}{},
	}
}
