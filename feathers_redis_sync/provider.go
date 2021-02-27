package feathers_redis_sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/tobiasbeck/feathers-go/feathers"
)

type RedisPublishMessage struct {
	Room    string      `json:"room"`
	Message interface{} `json:"message"`
}

//RedisSync is a provider which syncs realtime events over multiple server instances
type RedisSync struct {
	app    *feathers.App
	client *redis.Client
}

func (rs *RedisSync) Listen(port int, mux *http.ServeMux) {
	go func() {
		pubsub := rs.client.Subscribe("created", "removed", "patched", "updated")
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				panic(err)
			}
			var data RedisPublishMessage
			err = json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				fmt.Println("REDIS ERROR", err.Error())
				continue
			}
			rs.app.PublishToProviders(data.Room, msg.Channel, data.Message, "redis-sync")
		}
	}()
}

func (rs *RedisSync) Publish(room string, event string, data interface{}, provider string) {
	if provider == "redis-sync" {
		// If this has send the publish event skip
		return
	}
	message := RedisPublishMessage{
		Room:    room,
		Message: data,
	}
	encoded, err := json.Marshal(message)
	if err == nil {
		rs.client.Publish(event, encoded)
	}
}

// Configures a new RedisSync provider which synchronizes events for having multiple instances of server
func ConfigureRedisSync(app *feathers.App, config map[string]interface{}) error {
	app.Configure(configureRedisClient, config)
	if client, ok := app.Config("redisClient"); ok {
		provider := &RedisSync{
			app:    app,
			client: client.(*redis.Client),
		}
		app.AddProvider("redis-sync", provider)
		return nil
	}

	return errors.New("redis client not properly configured")
}
