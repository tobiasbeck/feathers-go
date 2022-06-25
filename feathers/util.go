package feathers

import (
	"reflect"

	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
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

type listenerEntry struct {
	key      uuid.UUID
	listener EventListener
	once     bool
}

type EventListener = func(data interface{})

type EventEmitter struct {
	eventListeners map[string][]listenerEntry
}

func (el *EventEmitter) Emit(event string, data interface{}) {
	if listeners, ok := el.eventListeners[event]; ok {
		nl := make([]listenerEntry, 0, len(listeners))
		for _, listener := range listeners {
			listener.listener(data)
			if !listener.once {
				nl = append(nl, listener)
			}
		}
		el.eventListeners[event] = nl
	}
}

type EventListenerUnregister = func() bool

func (el *EventEmitter) On(event string, listener EventListener) EventListenerUnregister {
	if _, ok := el.eventListeners[event]; !ok {
		el.eventListeners[event] = make([]listenerEntry, 0)
	}
	id, _ := uuid.New()
	listenerE := listenerEntry{
		key:      id,
		listener: listener,
		once:     false,
	}
	el.eventListeners[event] = append(el.eventListeners[event], listenerE)
	return func() bool {
		for k, listener := range el.eventListeners[event] {
			if uuid.Equal(listener.key, listenerE.key) {
				el.eventListeners[event] = append(el.eventListeners[event][:k], el.eventListeners[event][k+1:]...)
				return true
			}
		}
		return false
	}
}

func (el *EventEmitter) Once(event string, listener EventListener) {
	if _, ok := el.eventListeners[event]; !ok {
		el.eventListeners[event] = make([]listenerEntry, 0)
	}
	listenerE := listenerEntry{
		listener: listener,
		once:     true,
	}
	el.eventListeners[event] = append(el.eventListeners[event], listenerE)
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		eventListeners: make(map[string][]listenerEntry),
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
