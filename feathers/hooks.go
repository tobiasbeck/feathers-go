package feathers

import "context"

type CallMethod string

func (c CallMethod) String() string {
	return string(c)
}

type HookType string

func (c HookType) String() string {
	return string(c)
}

const (
	Find   CallMethod = "find"
	Get    CallMethod = "get"
	Create CallMethod = "create"
	Update CallMethod = "update"
	Patch  CallMethod = "patch"
	Remove CallMethod = "remove"
)

func eventFromCallMethod(method CallMethod) string {
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
	Before HookType = "before"
	After  HookType = "after"
	Error  HookType = "error"
)

// HookParams is the params passed to functions and go-feathers hooks
type HookParams struct {
	Params            map[string]interface{}
	Provider          string
	Route             string
	Headers           string
	CallContext       context.Context
	CallContextCancel context.CancelFunc
	fields            map[string]interface{}
}

func (hc *HookParams) GetField(key string) (interface{}, bool) {
	value, ok := hc.fields[key]
	return value, ok
}

func (hc *HookParams) SetField(key string, value interface{}) {
	hc.fields[key] = value
}

// HookContext is the context which is passed to every go-feathers hook
type HookContext struct {
	App        App
	Data       interface{}
	Error      error
	ID         string
	Method     CallMethod
	Path       string
	Result     interface{}
	Service    interface{}
	StatusCode int
	Type       HookType

	Params HookParams
}

type Hook = func(ctx *HookContext) (*HookContext, error)
type BoolHook = func(ctx *HookContext) (bool, error)
