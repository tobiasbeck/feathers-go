package feathers_mongo

import (
	"context"
	"errors"
	"time"

	"github.com/tobiasbeck/feathers-go/feathers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConfigureMongoClient(app *feathers.App, config map[string]interface{}) error {

	if mongodb, ok := app.GetConfig("mongodb"); ok {
		if mongoConfig, ok := mongodb.(map[interface{}]interface{}); ok {
			if uri, ok := mongoConfig["uri"]; ok {
				if db, ok := mongoConfig["db"]; ok {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri.(string)))
					if err != nil {
						return err
					}
					app.SetConfig("mongoClient", &client)
					app.SetConfig("mongoDb", client.Database(db.(string)))
					return nil
				}
				return errors.New("mongodb.db not found in config")
			}
			return errors.New("mongodb.uri not found in config")
		}
		return errors.New("could not parse mongodb configuration from config")
	}
	return errors.New("mongodb config is not set in config file")
}
