package feathers

import (
	"context"
	"errors"
	"strings"

	"github.com/mcuadros/go-lookup"
	"github.com/mitchellh/mapstructure"
)

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

type NewParamsOpt = func(params *Params)

func WithQuery(query Query) NewParamsOpt {
	return func(params *Params) {
		params.Query = query
	}
}

func WithOption(option string, value interface{}) NewParamsOpt {
	return func(params *Params) {
		params.Set(option, value)
	}
}

func NewParams(opts ...NewParamsOpt) *Params {
	params := &Params{
		Params: make(map[string]interface{}),
		fields: make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

func NewParamsFrom(origParams *Params, opts ...NewParamsOpt) *Params {
	params := origParams.Copy()
	for _, opt := range opts {
		opt(&params)
	}
	return &params
}
