package feathers_auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
	defaults "github.com/mcuadros/go-defaults"
	lookup "github.com/mcuadros/go-lookup"
	"github.com/mitchellh/mapstructure"
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/feathers_error"
)

type basePayload struct {
	Issuer         string       `json:"iss,omitempty" mapstructure:"issuer,omitempty"`
	Subject        string       `json:"sub,omitempty" mapstructure:"subject,omitempty"`
	Audience       jwt.Audience `json:"aud,omitempty" mapstructure:"audience,omitempty"`
	ExpirationTime *jwt.Time    `json:"exp,omitempty"`
	NotBefore      *jwt.Time    `json:"nbf,omitempty"`
	IssuedAt       *jwt.Time    `json:"iat,omitempty"`
	JWTID          string       `json:"jti,omitempty"`
}

type jwtToken struct {
	Payload basePayload
}

type AuthService struct {
	*feathers.BaseService
	*feathers.ModelService
	app            *feathers.App
	encryption     *jwt.HMACSHA
	config         map[string]interface{}
	authStrategies map[string]AuthStrategy
}

func tokenType(tpe string) jwt.SignOption {
	return func(hd *jwt.Header) {
		hd.Type = tpe
	}
}

func convertConfiguration(data interface{}) map[string]interface{} {

	converted := make(map[string]interface{})

	for key, value := range data.(map[interface{}]interface{}) {
		switch key := key.(type) {
		case string:
			converted[key] = value
		}
	}
	return converted
}

func (as *AuthService) Create(data map[string]interface{}, params feathers.Params) (interface{}, error) {
	model := Model{}
	err := as.MapAndValidateStruct(data, &model)
	if err != nil {
		return nil, feathers_error.Convert(err)
	}
	if strategy, ok := as.authStrategies[model.Strategy]; ok {
		result, err := strategy.Authenticate(model, params)
		if err != nil {
			return nil, feathers_error.Convert(err)
		}
		if _, ok := result["accessToken"]; ok {
			return result, nil
		}
		token, decoded, err := as.createAccessToken(result)
		if err != nil {
			return nil, feathers_error.Convert(err)
		}
		result["accessToken"] = token
		result["authentication"] = map[string]interface{}{
			"accessToken": token,
			"payload":     decoded,
		}
		return result, nil
	}
	return nil, feathers_error.NewGeneralError("Strategy "+model.Strategy+" not registered", nil)
}

// Remove TODO Add implementation
func (as *AuthService) Remove(id string, params feathers.Params) (interface{}, error) {
	return nil, feathers_error.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Find(params feathers.Params) (interface{}, error) {
	return nil, feathers_error.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Get(id string, params feathers.Params) (interface{}, error) {
	return nil, feathers_error.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Patch(id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	return nil, feathers_error.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Update(id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	return nil, feathers_error.NewMethodNotAllowed("Not supported", nil)
}

func Configure(app *feathers.App, config map[string]interface{}) error {
	if strategies, ok := config["strategies"]; ok {
		service := &AuthService{
			BaseService:    &feathers.BaseService{},
			ModelService:   feathers.NewModelService(NewModel),
			authStrategies: strategies.(map[string]AuthStrategy),
		}
		if config, ok := app.Config("authentication"); ok {
			convertedConfig := convertConfiguration(config)
			service.config = convertedConfig
			fmt.Printf("Config: %#v\n", convertedConfig)
			for key, strategy := range service.authStrategies {
				strategy.SetConfiguration(convertedConfig)
				strategy.SetApp(app)
				strategy.SetName(key)
			}
			service.encryption = jwt.NewHS256([]byte(service.config["secret"].(string)))
		} else {
			return errors.New("No app configuration of auth is set")
		}
		app.AddService("authentication", service)
		return nil
	}
	return errors.New("Strategies config not passed")
}

func (as *AuthService) DefaultConfig() DefaultAuthConfig {
	config := DefaultAuthConfig{}
	mapstructure.Decode(as.config, &config)
	defaults.SetDefaults(&config)
	return config
}

func (as *AuthService) createAccessToken(payload interface{}) (string, *jwtToken, error) {
	now := time.Now()
	defaultConfig := as.DefaultConfig()
	if entityKey, err := lookup.LookupString(payload, defaultConfig.Entity+"._id"); err != nil {
		if jwtConfig, ok := as.config["jwtOptions"]; ok {
			defaultPayload := basePayload{
				ExpirationTime: jwt.NumericDate(now.Add(24 * 30 * 12 * time.Hour)),
				NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
				IssuedAt:       jwt.NumericDate(now),
				Subject:        entityKey.String(),
				JWTID:          Uuid4(),
			}
			mapstructure.Decode(jwtConfig, &defaultPayload)
			pl := jwtToken{
				Payload: defaultPayload,
			}
			var tkTypeS string
			if tkType, err := lookup.LookupString(jwtConfig, "header.typ"); err != nil {
				tkTypeS = "access"
			} else {
				tkTypeS = tkType.String()
			}
			token, err := jwt.Sign(pl, as.encryption, tokenType(tkTypeS))
			if err != nil {
				return "", nil, err
			}
			return string(token), &pl, nil
		}
		return "", nil, errors.New("No jwtOptions found")
	} else {
		return "", nil, err
	}
}

// const authStrategies = params.authStrategies || this.configuration.authStrategies;

// if (!authStrategies.length) {
// 	throw new NotAuthenticated('No authentication strategies allowed for creating a JWT (`authStrategies`)');
// }

// const authResult = await this.authenticate(data, params, ...authStrategies);

// debug('Got authentication result', authResult);

// if (authResult.accessToken) {
// 	return authResult;
// }

// const [ payload, jwtOptions ] = await Promise.all([
// 	this.getPayload(authResult, params),
// 	this.getTokenOptions(authResult, params)
// ]);

// debug('Creating JWT with', payload, jwtOptions);

// const accessToken = await this.createAccessToken(payload, jwtOptions, params.secret);

// return merge({ accessToken }, authResult, {
// 	authentication: {
// 			accessToken,
// 			payload: jsonwebtoken.decode(accessToken)
// 	}
// });
