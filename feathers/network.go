package feathers

// Caller represents a caller of a request. It handles Callbacks
type Caller interface {
	Callback(data interface{})
	CallbackError(err error)
}
