package feathers_auth

import (
	"context"
	"errors"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/tobiasbeck/feathers-go/feathers"
)

type JwtStrategy struct {
	*BaseAuthStrategy
}

func (s *JwtStrategy) Authenticate(ctx context.Context, data Model, params feathers.Params) (map[string]interface{}, error) {
	defaultConfig := s.DefaultConfig()
	if jwtConfig, ok := s.config["jwtOptions"]; ok {
		if token, ok := data.Params["accessToken"]; ok {
			encryption := jwt.NewHS256([]byte(defaultConfig.Secret))
			var payload jwtToken
			now := time.Now()

			defaultPayload := jwtToken{}

			mapstructure.Decode(jwtConfig, &defaultPayload)
			defaults.SetDefaults(&defaultPayload)

			iatValidator := jwt.IssuedAtValidator(now)
			expValidator := jwt.ExpirationTimeValidator(now)
			// audValidator := jwt.AudienceValidator(defaultPayload.Audience)
			issValidator := jwt.IssuerValidator(defaultPayload.Issuer)
			// nbfValidator := jwt.NotBeforeValidator(now)

			validatePayload := jwt.ValidatePayload((*jwt.Payload)(&payload), iatValidator, expValidator, issValidator)
			_, err := jwt.Verify([]byte(token.(string)), encryption, &payload, validatePayload)
			if err != nil {
				return nil, err
			}

			entityID := payload.Subject

			if entityService, ok := s.EntityService(); ok {
				entity, err := entityService.Get(ctx, entityID, *feathers.NewParams())
				if err != nil {
					return nil, err
				}
				result := map[string]interface{}{
					"authentication": map[string]interface{}{
						"strategy":    "local",
						"accessToken": token,
						"payload":     payload,
					},
				}
				result[defaultConfig.Entity] = entity
				result["accessToken"] = token
				return result, nil

			}
			return nil, errors.New("JWT Invalid")
		}
		return nil, errors.New("No accessToken sent")
	}
	return nil, errors.New("no jwt configuration given")
}

func NewJwtStrategy() *JwtStrategy {
	return &JwtStrategy{
		BaseAuthStrategy: &BaseAuthStrategy{},
	}
}
