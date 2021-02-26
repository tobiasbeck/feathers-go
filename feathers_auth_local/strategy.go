package feathers_auth_local

import "github.com/tobiasbeck/feathers-go/feathers_auth"

type Strategy struct {
	*feathers_auth.BaseAuthStrategy
}

func New() *Strategy {
	return &Strategy{
		BaseAuthStrategy: &feathers_auth.BaseAuthStrategy{},
	}
}
