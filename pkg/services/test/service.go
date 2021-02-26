package service_test

import (
	"github.com/tobiasbeck/hackero/pkg/feathers"
	feathersmongo "github.com/tobiasbeck/hackero/pkg/feathers-mongo"
)

type TestService interface {
}

type testservice struct {
	*feathers.BasePublishableService
	*feathersmongo.Service
	Hooks feathers.HooksTree
}
