package redis

import (
	"errors"

	"github.com/go-redis/redis"
	"github.com/tobiasbeck/feathers-go/feathers"
)

func configureRedisClient(app *feathers.App, config map[string]interface{}) error {
	if rd, ok := app.Config("redis"); ok {
		if redisConfig, ok := rd.(map[string]interface{}); ok {
			if addr, ok := redisConfig["address"]; ok {
				client := redis.NewClient(&redis.Options{
					Addr:     addr.(string),
					Password: "",
					DB:       0,
				})
				app.SetConfig("redisClient", client)
				return nil
			}
			return errors.New("redis.address not found")
		}
		return errors.New("could not parse redis config")
	}
	return errors.New("redis config is not set in config file")
}
