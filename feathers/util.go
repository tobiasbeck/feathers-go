package feathers

import (
	"reflect"
)

func getField(v interface{}, field string) (interface{}, bool) {
	r := reflect.ValueOf(v)
	if !r.Elem().FieldByName(field).IsValid() {
		return nil, false
	}
	f := reflect.Indirect(r).FieldByName(field)
	return f.Interface(), true
}

type EventListener = func(data interface{})

type EventEmitter struct {
	eventListeners map[string][]EventListener
}

func (el *EventEmitter) Emit(event string, data interface{}) {
	if listeners, ok := el.eventListeners[event]; ok {
		for _, listener := range listeners {
			listener(data)
		}
	}
}

func (el *EventEmitter) On(event string, listener EventListener) {
	if _, ok := el.eventListeners[event]; !ok {
		el.eventListeners[event] = make([]EventListener, 0)
	}
	el.eventListeners[event] = append(el.eventListeners[event], listener)
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		eventListeners: make(map[string][]EventListener),
	}
}
