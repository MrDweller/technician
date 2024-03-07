package event

import (
	"fmt"
)

type Subscriber struct {
	listeners map[Event]Listener
}

func NewSubscriber() *Subscriber {
	listeners := map[Event]Listener{}
	return &Subscriber{
		listeners: listeners,
	}
}

func (subscriber *Subscriber) Subscribe(address string, port int, event Event, output chan<- []byte) error {
	_, exists := subscriber.listeners[event]
	if exists {
		return fmt.Errorf("alredy subscribed to %s event", event.Name)
	}
	subscriber.listeners[event] = NewRabbitmqListener()
	return subscriber.listeners[event].Listen(address, port, event, output)
}

func (subscriber *Subscriber) Unsubscribe(event Event) error {
	listener, exists := subscriber.listeners[event]
	if !exists {
		return fmt.Errorf("not subscribed to %s event", event.Name)
	}
	delete(subscriber.listeners, event)
	return listener.Stop()
}
