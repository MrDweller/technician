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

func (subscriber *Subscriber) Subscribe(address string, port int, event Event, metadata map[string]string, output chan<- []byte) error {
	_, exists := subscriber.listeners[event]
	if exists {
		return fmt.Errorf("already subscribed to %s event", event.Name)
	}
	subscriber.listeners[event] = NewRabbitmqListener()
	return subscriber.listeners[event].Listen(address, port, event, metadata, output)
}

func (subscriber *Subscriber) Unsubscribe(event Event) error {
	listener, exists := subscriber.listeners[event]
	if !exists {
		return fmt.Errorf("not subscribed to %s event", event.Name)
	}
	delete(subscriber.listeners, event)
	return listener.Stop()
}

func (subscriber Subscriber) UnsubscribeAll() error {
	var err error
	err = nil
	for event := range subscriber.listeners {
		subErr := subscriber.Unsubscribe(event)
		if subErr != nil {
			err = subErr
		}
	}

	return err
}
