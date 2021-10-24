package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/tobiasbeck/feathers-go/feathers"
)

type FeathersMessageContext struct {
	Method feathers.RestMethod `json:"method"`
	ID     string              `json:"id"`
	Type   feathers.HookType   `json:"type"`
}
type FeathersMessage struct {
	Context FeathersMessageContext `json:"context"`
	Event   string                 `json:"event"`
	Data    interface{}            `json:"data"`
	Path    string                 `json:"path"`
}
type RedisPublishMessage struct {
	Room    string      `json:"room"`
	Path    string      `json:"path"`
	Message interface{} `json:"message"`
}

//RedisSync is a provider which syncs realtime events over multiple server instances
type RedisSync struct {
	app    *feathers.App
	client *redis.Client
}

func (rs *RedisSync) Listen(port int, mux *http.ServeMux) {
	go func() {
		pubsub := rs.client.Subscribe("created", "removed", "patched", "updated", "feathers-sync")
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				fmt.Println("REDIS ERROR", err.Error())
				continue
			}
			if msg.Channel == "feathers-sync" {
				var data FeathersMessage
				err = json.Unmarshal([]byte(msg.Payload), &data)
				if err != nil {
					fmt.Println("REDIS ERROR", err.Error())
					continue
				}
				// TODO: Better context
				params := feathers.NewParams()
				params.Provider = "redis-sync"
				service := rs.app.Service(data.Path)
				serviceClass := rs.app.ServiceClass(data.Path)
				triggerContext := &feathers.Context{
					App:          *rs.app,
					Data:         data.Data.(map[string]interface{}),
					Result:       data.Data,
					Method:       data.Context.Method,
					Path:         data.Path,
					ID:           data.Context.ID,
					Service:      service,
					ServiceClass: serviceClass,
					Type:         data.Context.Type,
					Params:       *params,
				}

				rs.app.TriggerUpdate(triggerContext)
			} else {
				var data RedisPublishMessage
				err = json.Unmarshal([]byte(msg.Payload), &data)
				if err != nil {
					fmt.Println("REDIS ERROR", err.Error())
					continue
				}
				rs.app.PublishToProviders(data.Room, msg.Channel, data.Message, data.Path, "redis-sync")
			}
		}
	}()
}

func (rs *RedisSync) Publish(room string, event string, data interface{}, path string, provider string) {
	if provider == "redis-sync" {
		// If this has send the publish event skip
		return
	}
	message := RedisPublishMessage{
		Room:    room,
		Message: data,
		Path:    path,
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
