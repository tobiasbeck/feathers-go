package feathers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mcuadros/go-lookup"
	"github.com/mitchellh/mapstructure"
)

func deepCopyMap(origMap map[string]interface{}) map[string]interface{} {
	nM := map[string]interface{}{}
	for key, value := range origMap {
		switch v := value.(type) {
		case map[string]interface{}:
			nM[key] = deepCopyMap(v)
		default:
			nM[key] = v
		}
	}
	return nM
}

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

type Query = map[string]interface{}

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
	fields  map[string]interface{}
	// Query conatains query fields specified by client
	Query Query

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

func (hc *Params) Has(key string) bool {
	_, ok := hc.fields[key]
	return ok
}

// Set sets a hook field (e.g. user, additional information etc.)
func (hc *Params) Set(key string, value interface{}) {
	if hc.fields == nil {
		hc.fields = map[string]interface{}{}
	}
	hc.fields[key] = value
}

func (hc *Params) New() Params {
	return *NewParams()
}

func (hc *Params) Copy() Params {
	np := *NewParams()
	np.Provider = hc.Provider
	np.Route = hc.Route
	np.Connection = hc.Connection
	np.Query = hc.Query
	np.User = hc.User
	np.Authenticated = hc.Authenticated
	np.Params = deepCopyMap(hc.Params)
	// np.Headers = deepCopyMap(hc.Headers)
	np.fields = deepCopyMap(hc.fields)
	return np
}

func (hc *Params) NewWithQuery(query map[string]interface{}) Params {
	return *NewParamsQuery(query)
}

// NewParams creates a empty params struct
func NewParams() *Params {
	return &Params{
		Params: make(map[string]interface{}),
		fields: make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}
}

func NewAuthenticatedParams(user map[string]interface{}) *Params {
	p := NewParams()
	p.Authenticated = true
	p.User = user
	return p
}

// NewParamsQuery returns a new HookParms struct only containng specified query
func NewParamsQuery(query map[string]interface{}) *Params {

	return &Params{
		Params: make(map[string]interface{}),
		fields: make(map[string]interface{}),
		Query:  query,
	}
}

type Data = map[string]interface{}

// func (d Data) Get(path []string) interface{} {
// 	value, err := lookup.Lookup(d, path...)
// 	if err != nil {
// 		return nil
// 	}
// 	return value.Interface()
// }

// Context is the context which is passed to every go-feathers hook (Can be used as context.Context)
type Context struct {
	context.Context
	// App is a reference to the current application instance
	App App
	// Data is the data passed from the requesting instance
	Data Data
	// Error contains the error which was triggered while executing the route
	Error error
	// ID is the id passed fromt the requesting instance
	ID string
	// Method containts the method which was called (Get, Patch, Find, etc.)
	Method RestMethod
	// Path is the path to the service
	Path string
	// Result is the result of the service call (only defined in after hooks)
	Result interface{}
	// Service is the called service, but wrapped by the application to also trigger hooks (If you wanna call service class directly for custom methods use ServiceClass)
	Service Service
	// ServiceClass is the current service without the wrapper. Do NOT Call Patch, Find, Get, Remove and Update since they do not call hooks!
	ServiceClass interface{}
	StatusCode   int
	// Type of the hook (Before or after)
	Type HookType
	// Params for this call
	Params Params
}

// DataMerge merges new data with already exsting data
func (c *Context) DataMerge(data Data) {
	for key, value := range data {
		c.Data[key] = value
	}
}

// DataDecode decodes data at `path` to target (pointer)
func (c *Context) DataDecode(target interface{}, path ...string) error {
	var data interface{} = c.Data
	if len(path) > 0 {
		data = c.DataGet(path...)
	}
	if data == nil {
		return errors.New("Data at path '" + strings.Join(path, ".") + "' not defined")
	}
	err := mapstructure.WeakDecode(data, target)
	if err != nil {
		return err
	}
	return nil
}
func (c *Context) DataGet(key ...string) interface{} {
	val, ok := lookup.Lookup(c.Data, key...)
	if ok != nil {
		return nil
	}
	return val.Interface()
}

func (c *Context) DataHas(key ...string) bool {
	val, ok := lookup.Lookup(c.Data, key...)
	if ok == nil && !val.IsZero() {
		return true
	}
	return false
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *Context) Err() error {
	return c.Context.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

// Hook is a function which can be used to modify request params
type Hook = func(ctx *Context) (*Context, error)

// BoolHook works just like a hook but returns a boolean and cannot modify the context
type BoolHook = func(ctx *Context) (bool, error)
