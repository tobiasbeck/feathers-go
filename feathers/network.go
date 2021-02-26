package feathers

type Caller interface {
	Callback(data interface{})
	CallbackError(err error)
}
