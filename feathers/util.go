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

func setField(v interface{}, field string, value interface{}) {
	r := reflect.ValueOf(v)
	rValue := reflect.ValueOf(value)
	if !r.Elem().FieldByName(field).IsValid() {
		return
	}
	reflect.Indirect(r).FieldByName(field).Set(rValue)
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

// StructToMap converts a model struct into an interface
func StructToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	decoder, err := newDecoder(&result)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return result, nil
}
