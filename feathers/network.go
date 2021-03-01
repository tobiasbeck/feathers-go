package feathers

// Caller represents a caller of a request. It handles Callbacks
type Caller interface {
	Callback(data interface{})
	CallbackError(err error)
	IsSocket() bool
	SocketConnection() Connection
}

type Connection interface {
	Join(room string) error
	Leave(room string) error
	IsAuthenticated() bool
	AuthEntity() interface{}
}
