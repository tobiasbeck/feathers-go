package feathers

import (
	"reflect"
	"sync"

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
	key     uuid.UUID
	channel chan interface{}
	once    bool
}

type topic struct {
	listeners []listenerEntry
	sync.RWMutex
}

type EventEmitter struct {
	eventListeners map[string]*topic
}

func (el *EventEmitter) Emit(event string, data interface{}) {
	if eventTopic, ok := el.eventListeners[event]; ok {
		eventTopic.RLock()
		nl := make([]listenerEntry, 0, len(eventTopic.listeners))
		for _, listener := range eventTopic.listeners {
			listener.channel <- data
			if !listener.once {
				nl = append(nl, listener)
			} else {
				close(listener.channel)
			}
		}
		eventTopic.RUnlock()
		eventTopic.Lock()
		defer eventTopic.Unlock()
		t := el.eventListeners[event]
		t.listeners = nl
		el.eventListeners[event] = t
	}
}

func (el *EventEmitter) initTopic(topicName string) {
	if _, ok := el.eventListeners[topicName]; !ok {
		el.eventListeners[topicName] = &topic{
			listeners: make([]listenerEntry, 0),
		}
	}
}

type EventListenerUnregister = func() bool

func (el *EventEmitter) On(event string) (<-chan interface{}, EventListenerUnregister) {
	if _, ok := el.eventListeners[event]; !ok {
		el.initTopic(event)
	}
	id, _ := uuid.New()
	listenerE := listenerEntry{
		key:     id,
		channel: make(chan interface{}),
		once:    false,
	}

	eventTopic := el.eventListeners[event]
	eventTopic.Lock()
	defer eventTopic.Unlock()
	eventTopic.listeners = append(eventTopic.listeners, listenerE)
	el.eventListeners[event] = eventTopic
	return listenerE.channel, func() bool {
		eventTopic.Lock()
		defer eventTopic.Unlock()
		for k, listener := range eventTopic.listeners {
			if uuid.Equal(listener.key, listenerE.key) {
				eventTopic.listeners = append(eventTopic.listeners[:k], eventTopic.listeners[k+1:]...)
				return true
			}
		}
		return false
	}
}

func (el *EventEmitter) Once(event string) <-chan interface{} {
	if _, ok := el.eventListeners[event]; !ok {
		el.initTopic(event)
	}
	listenerE := listenerEntry{
		channel: make(chan interface{}),
		once:    true,
	}
	eventTopic := el.eventListeners[event]
	eventTopic.Lock()
	defer eventTopic.Unlock()
	eventTopic.listeners = append(eventTopic.listeners, listenerE)
	el.eventListeners[event] = eventTopic
	return listenerE.channel
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		eventListeners: make(map[string]*topic),
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
