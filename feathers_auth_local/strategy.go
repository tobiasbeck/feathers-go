package feathers_auth_local

import (
	"fmt"

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

func (s *Strategy) findEntity(username string, params feathers.Params) (map[string]interface{}, error) {
	config := strategyConfig{}
	s.StrategyConfig(&config)
	if service, ok := s.EntityService(); ok {
		query := map[string]interface{}{}
		query[config.UsernameField] = username
		params := feathers.NewParamsQuery(query)
		iResults, err := service.Find(*params)
		if err != nil {
			return nil, feathers_error.NewNotAuthenticated(err.Error(), nil)
		}
		results := iResults.([]map[string]interface{})
		if len(results) >= 1 {
			fmt.Printf("entity %#v\n", results[0])
			return results[0], nil
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

func (s *Strategy) Authenticate(data feathers_auth.Model, params feathers.Params) (map[string]interface{}, error) {
	config := strategyConfig{}
	s.StrategyConfig(&config)
	defaultConfig := s.DefaultConfig()
	fmt.Printf("data: %s, %#v\n", config.UsernameField, data)
	user, err := s.findEntity(data.Params[config.UsernameField].(string), params)
	if err != nil {
		return nil, err
	}
	passwordCorrect := s.comparePassword(user, data.Params[config.PasswordField].(string))
	if !passwordCorrect {
		return nil, feathers_error.NewNotAuthenticated("Username or Password is incorrect", nil)
	}
	result := map[string]interface{}{
		"authentication": struct{ Strategy string }{Strategy: "local"},
	}
	result[defaultConfig.Entity] = user
	return result, nil
}

func New() *Strategy {
	return &Strategy{
		BaseAuthStrategy: &feathers_auth.BaseAuthStrategy{},
	}
}
