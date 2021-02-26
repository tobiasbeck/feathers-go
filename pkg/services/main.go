package services

import (
	service_test "github.com/tobiasbeck/hackero/app/services/test"

	"github.com/tobiasbeck/hackero/pkg/feathers"
)

func ConfigureServices(app *feathers.App, config map[string]interface{}) error {
	service_test.ConfigureService(app)
	return nil
}
