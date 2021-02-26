package app

import (
	"fmt"

	"github.com/tobiasbeck/hackero/pkg/feathers"
	gosocketio "github.com/tobiasbeck/hackero/pkg/gosf-socketio"
)

func ConfigureChannels(app *feathers.App, config map[string]interface{}) error {
	app.On("connection", func(data interface{}) {
		if channel, ok := data.(*gosocketio.Channel); ok {
			channel.Join("anonymous")
			fmt.Println("joined anonymous")
		}
	})
	return nil
}
