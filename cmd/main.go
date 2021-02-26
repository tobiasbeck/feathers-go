package main

import "github.com/tobiasbeck/feathers-go/feathers-go/feathers"

func main() {
	app := feathers.NewApp()
	app.LoadConfig("./config/default.yaml")
	app.Configure(feathersMongo.ConfigureMongoClient, make(map[string]interface{}))
	app.Configure(feathers.ConfigureHttpProvider, make(map[string]interface{}))
	app.Configure(feathers.ConfigureSocketIOProvider, make(map[string]interface{}))
	app.Configure(feathersRedisSync.ConfigureRedisSync, make(map[string]interface{}))
	app.Configure(services.ConfigureServices, make(map[string]interface{}))
	app.Configure(application.ConfigureChannels, make(map[string]interface{}))
	app.SetAppHooks(application.AppHooks)
	app.Startup(4)
	app.Listen()
}
