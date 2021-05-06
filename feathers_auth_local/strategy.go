package feathers_auth_local

import (
	"context"

	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
	"github.com/tobiasbeck/feathers-go/feathers_auth"
	"golang.org/x/crypto/bcrypt"
)

type strategyConfig struct {
	UsernameField string `mapstructure:"usernameField" default:"username"`
	PasswordField string `mapstructure:"passswordField" default:"password"`
}

type Strategy struct {
	*feathers_auth.BaseAuthStrategy
}

func (s *Strategy) findEntity(ctx context.Context, username string, params feathers.Params) (map[string]interface{}, error) {
	config := strategyConfig{}
	s.StrategyConfig(&config)
	if service, ok := s.EntityService(); ok {
		query := map[string]interface{}{}
		query[config.UsernameField] = username
		findParams := feathers.NewParamsQuery(query)
		iResults, err := service.Find(ctx, *findParams)
		if err != nil {
			return nil, feathers_error.NewNotAuthenticated(err.Error(), nil)
		}
		// fmt.Printf("IRESULTS: %#v\n", iResults)
		switch result := iResults.(type) {
		case []map[string]interface{}:
			if len(result) > 0 {
				// fmt.Printf("entity %#v\n", result[0])
				return result[0], nil
			}
			return nil, feathers_error.NewNotAuthenticated("", nil)
		case map[string]interface{}:
			// fmt.Printf("entity %#v\n", result)
			return result, nil

		}
		return nil, feathers_error.NewNotAuthenticated("", nil)
	}
	return nil, feathers_error.NewNotAuthenticated("Not Authenticated", nil)
}

func (s *Strategy) comparePassword(entity map[string]interface{}, password string) bool {
	config := strategyConfig{}
	s.StrategyConfig(&config)
	if iEntPassword, ok := entity[config.PasswordField]; ok {
		entPassword := iEntPassword.(string)
		err := bcrypt.CompareHashAndPassword([]byte(entPassword), []byte(password))
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func (s *Strategy) Authenticate(ctx context.Context, data feathers_auth.Model, params feathers.Params) (map[string]interface{}, error) {
	config := strategyConfig{}
	s.StrategyConfig(&config)
	defaultConfig := s.DefaultConfig()
	// fmt.Printf("data: %s, %#v\n", config.UsernameField, data)
	entity, err := s.findEntity(ctx, data.Params[config.UsernameField].(string), params)
	if err != nil {
		return nil, err
	}
	// This takes around 250ms to complete (pretty slow)
	passwordCorrect := s.comparePassword(entity, data.Params[config.PasswordField].(string))
	if !passwordCorrect {
		return nil, feathers_error.NewNotAuthenticated("Username or Password is incorrect", nil)
	}
	result := map[string]interface{}{
		"authentication": struct{ Strategy string }{Strategy: "local"},
	}
	result[defaultConfig.Entity] = entity
	return result, nil
}

func New() *Strategy {
	return &Strategy{
		BaseAuthStrategy: &feathers_auth.BaseAuthStrategy{},
	}
}
