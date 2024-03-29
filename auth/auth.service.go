package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
	defaults "github.com/mcuadros/go-defaults"
	lookup "github.com/mcuadros/go-lookup"
	"github.com/mitchellh/mapstructure"
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/httperrors"
)

type hexable interface {
	Hex() string
}

type jwtToken struct {
	Issuer         string       `json:"iss,omitempty" mapstructure:"issuer,omitempty"`
	Subject        string       `json:"sub,omitempty" mapstructure:"subject,omitempty"`
	Audience       jwt.Audience `json:"aud,omitempty" mapstructure:"audience,omitempty"`
	ExpirationTime *jwt.Time    `json:"exp,omitempty"`
	NotBefore      *jwt.Time    `json:"nbf,omitempty"`
	IssuedAt       *jwt.Time    `json:"iat,omitempty"`
	JWTID          string       `json:"jti,omitempty"`
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

func (as *AuthService) Create(ctx context.Context, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	model := Model{}
	err := as.MapAndValidateStruct(data, &model)
	if err != nil {
		return nil, httperrors.Convert(err)
	}
	if strategy, ok := as.authStrategies[model.Strategy]; ok {
		result, err := strategy.Authenticate(ctx, model, params)
		if err != nil {
			return nil, httperrors.Convert(err)
		}

		if params.IsSocket && params.Connection != nil {
			defaultConfig := as.DefaultConfig()
			params.Connection.SetAuthEntity(result[defaultConfig.Entity])
			as.app.Emit("login", params.Connection)
		}

		if _, ok := result["accessToken"]; ok {
			return result, nil
		}
		token, decoded, err := as.createAccessToken(result)
		if err != nil {
			return nil, httperrors.Convert(err)
		}

		result["accessToken"] = token
		result["authentication"] = map[string]interface{}{
			"accessToken": token,
			"payload":     decoded,
		}

		return result, nil
	}
	return nil, httperrors.NewGeneralError("Strategy "+model.Strategy+" not registered", nil)
}

// Remove TODO Add implementation
func (as *AuthService) Remove(ctx context.Context, id string, params feathers.Params) (interface{}, error) {
	return nil, httperrors.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Find(ctx context.Context, params feathers.Params) (interface{}, error) {
	return nil, httperrors.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Get(ctx context.Context, id string, params feathers.Params) (interface{}, error) {
	return nil, httperrors.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Patch(ctx context.Context, id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	return nil, httperrors.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Update(ctx context.Context, id string, data map[string]interface{}, params feathers.Params) (interface{}, error) {
	return nil, httperrors.NewMethodNotAllowed("Not supported", nil)
}

func NewAuthService(app *feathers.App, strategies map[string]AuthStrategy) *AuthService {
	service := &AuthService{
		app:            app,
		BaseService:    &feathers.BaseService{},
		ModelService:   feathers.NewModelService(NewModel),
		authStrategies: strategies,
	}
	if appConfig, ok := app.Config("authentication"); ok {
		appMapConfig := appConfig.(map[string]interface{})
		service.config = appMapConfig
		for key, strategy := range service.authStrategies {
			strategy.SetConfiguration(appMapConfig)
			strategy.SetApp(app)
			strategy.SetName(key)
		}
		service.encryption = jwt.NewHS256([]byte(service.config["secret"].(string)))
	} else {
		panic("No app configuration of auth is set")
	}
	return service
}

func Configure(app *feathers.App, config map[string]interface{}) error {
	if strategies, ok := config["strategies"]; ok {
		strategyList := strategies.(map[string]AuthStrategy)
		service := NewAuthService(app, strategyList)
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
	if entityKey, err := lookup.LookupString(payload, defaultConfig.Entity+"._id"); err == nil {
		if jwtConfig, ok := as.config["jwtOptions"]; ok {

			var stringKey string
			switch key := entityKey.Interface().(type) {
			case string:
				stringKey = key
			case hexable:
				stringKey = key.Hex()
			case fmt.Stringer:
				stringKey = key.String()
			default:
				return "", nil, errors.New("Cannot strinigy entity key")
			}

			payload := jwtToken{
				ExpirationTime: jwt.NumericDate(now.Add(24 * time.Hour)),
				NotBefore:      jwt.NumericDate(now),
				IssuedAt:       jwt.NumericDate(now),
				Subject:        stringKey,
				JWTID:          Uuid4(),
			}
			mapstructure.Decode(jwtConfig, &payload)
			var tkTypeS string
			if tkType, err := lookup.Lookup(jwtConfig, "header.typ"); err != nil {
				tkTypeS = "access"
			} else {
				tkTypeS = tkType.String()
			}
			token, err := jwt.Sign(payload, as.encryption, tokenType(tkTypeS))
			if err != nil {
				return "", nil, err
			}
			return string(token), &payload, nil
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
