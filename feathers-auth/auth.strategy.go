package feathersAuth

import "github.com/tobiasbeck/feathers-go/feathers"

type AuthStrategy interface {
	Authenticate(data Model, params feathers.HookParams) (map[string]interface{}, error)
}
