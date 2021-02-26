package feathersAuth

import (
	"github.com/tobiasbeck/feathers-go/feathers"
	"github.com/tobiasbeck/feathers-go/feathers/fErr"
)

type AuthService struct {
	*feathers.BaseService
	*feathers.ModelService
	app            *feathers.App
	authStrategies map[string]AuthStrategy
}

func (as *AuthService) Create(data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	model := Model{}
	err := as.MapAndValidateStruct(data, &model)
	if err != nil {
		return nil, fErr.Convert(err)
	}
	if strategy, ok := as.authStrategies[model.Strategy]; ok {
		result, err := strategy.Authenticate(model, params)
		if err != nil {
			return nil, fErr.Convert(err)
		}
		if _, ok := result["accessToken"]; ok {
			return result, nil
		}
		return nil, fErr.NewGeneralError("Internal auth error occurred", nil)
	}
	return nil, fErr.NewGeneralError("Strategy "+model.Strategy+" not registered", nil)
}

// Remove TODO
func (as *AuthService) Remove(id string, params feathers.HookParams) (interface{}, error) {
	return nil, fErr.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Find(params feathers.HookParams) (interface{}, error) {
	return nil, fErr.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Get(id string, params feathers.HookParams) (interface{}, error) {
	return nil, fErr.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Patch(id string, data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	return nil, fErr.NewMethodNotAllowed("Not supported", nil)
}

func (as *AuthService) Update(id string, data map[string]interface{}, params feathers.HookParams) (interface{}, error) {
	return nil, fErr.NewMethodNotAllowed("Not supported", nil)
}

func ConfigureAuthentication(app *feathers.App, config map[string]interface{}) error {
	if strategies, ok := config["strategies"]; ok {
		service := &AuthService{
			BaseService:    &feathers.BaseService{},
			ModelService:   feathers.NewModelService(NewModel),
			authStrategies: strategies.(map[string]AuthStrategy),
		}
		app.AddService("authentication", service)
		return nil
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
