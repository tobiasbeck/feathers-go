package service_test

import (
	"github.com/tobiasbeck/hackero/pkg/feathers"
	feathersmongo "github.com/tobiasbeck/hackero/pkg/feathers-mongo"
)

func ConfigureService(app *feathers.App) error {
	mongoService := feathersmongo.NewService("test", NewTestModel, app)
	mongoService.BaseService.Hooks = serviceHooks
	service := &testservice{
		Service:                mongoService,
		BasePublishableService: &feathers.BasePublishableService{},
	}
	app.AddService("test", service)
	return nil
}
