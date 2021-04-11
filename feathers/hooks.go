package feathers

import (
	"context"

	"github.com/mcuadros/go-lookup"
)

// RestMethod represents the method which was called
type RestMethod string

func (c RestMethod) String() string {
	return string(c)
}

// HookType is either `Before, After or Error` and describes the current type of the hook chain
type HookType string

func (c HookType) String() string {
	return string(c)
}

const (
	// Find method retrieves multiple documents
	Find RestMethod = "find"
	// Get method retrieves a single document
	Get RestMethod = "get"
	// Create method creates a new document
	Create RestMethod = "create"
	// Update method replaces a whole doucmnet
	Update RestMethod = "update"
	// Patch method inserts new keys or updates existing keys
	Patch RestMethod = "patch"
	// Remove method removes a document
	Remove RestMethod = "remove"
)

func eventFromCallMethod(method RestMethod) string {
	switch method {
	case Create:
		return "created"
	case Update:
		return "updated"
	case Patch:
		return "patched"
	case Remove:
		return "removed"
	}
	return ""
}

const (
	// Before Hooks are executed before the service method
	Before HookType = "before"
	// After Hooks are executed after service method
	After HookType = "after"
	// Error Hooks are executed if a hook or service method returns a error
	Error HookType = "error"
)

// Params is the params passed to functions and go-feathers hooks
type Params struct {
	Params map[string]interface{}
	// Name of provider from which is called from (empty string for server)
	Provider string
	// Route which is called (service name)
	Route string
	//Caller instance of who has called this
	Connection Connection
	// True if connection is socket based
	IsSocket bool
	// Headers from client call
	Headers map[string]string
	// CallContext is a context passed through the whole execution. use this to derive your own contexts or pass it to calls requiring context
	CallContext context.Context
	// CallContextCancel is the cancel function of CallContext (called by system)
	CallContextCancel context.CancelFunc
	fields            map[string]interface{}
	// Query conatains query fields specified by client
	Query map[string]interface{}

	User map[string]interface{}

	Authenticated bool
}

// Get retrieves a field from the hooks
func (hc *Params) Get(key string) interface{} {
	value, ok := hc.fields[key]
	if !ok {
		return nil
	}
	return value
}

// Get retrieves a field from the hooks
func (hc *Params) Lookup(key string) (interface{}, bool) {
	value, ok := hc.fields[key]
	return value, ok
}

// Set sets a hook field (e.g. user, additional information etc.)
func (hc *Params) Set(key string, value interface{}) {
	if hc.fields == nil {
		hc.fields = map[string]interface{}{}
	}
	hc.fields[key] = value
}

// WithQuery creates new Params with given query only passing same CallContext
func (p *Params) WithQuery(query map[string]interface{}) *Params {
	return NewParamsQuery(p.CallContext, query)
}

// WithQuery creates new Params only passing same CallContext
func (p *Params) WithContext() *Params {
	return NewParams(p.CallContext)
}

// NewParams creates a empty params struct
func NewParams(ctx context.Context) *Params {
	callContext, cancel := context.WithCancel(ctx)
	return &Params{
		CallContext:       callContext,
		CallContextCancel: cancel,
		Params:            make(map[string]interface{}),
		fields:            make(map[string]interface{}),
		Query:             make(map[string]interface{}),
	}
}

// NewParamsQuery returns a new HookParms struct only containng specified query
func NewParamsQuery(ctx context.Context, query map[string]interface{}) *Params {
	callContext, cancel := context.WithCancel(ctx)
	return &Params{
		CallContext:       callContext,
		CallContextCancel: cancel,
		Params:            make(map[string]interface{}),
		fields:            make(map[string]interface{}),
		Query:             query,
	}
}

type Data map[string]interface{}

func (d Data) Get(path []string) interface{} {
	value, err := lookup.Lookup(d, path...)
	if err != nil {
		return nil
	}
	return value.Interface()
}

// Context is the context which is passed to every go-feathers hook
type Context struct {
	App        App
	Data       Data
	Error      error
	ID         string
	Method     RestMethod
	Path       string
	Result     interface{}
	Service    interface{}
	StatusCode int
	Type       HookType

	Params Params
}

// Hook is a function which can be used to modify request params
type Hook = func(ctx *Context) (*Context, error)

// BoolHook works just like a hook but returns a boolean and cannot modify the context
type BoolHook = func(ctx *Context) (bool, error)
