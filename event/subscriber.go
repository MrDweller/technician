package event

import (
	"fmt"
)

type Subscriber struct {
	listeners map[string]Listener
}

func NewSubscriber() *Subscriber {
	listeners := map[string]Listener{}
	return &Subscriber{
		listeners: listeners,
	}
}

func (subscriber *Subscriber) Subscribe(address string, port int, event Event, metadata map[string]string, output chan<- []byte) error {
	listener := NewRabbitmqListener(address, port, event, metadata)
	_, exists := subscriber.listeners[listener.GetListenerId()]
	if exists {
		return fmt.Errorf("already subscribed to %s event", event.Name)
	}
	subscriber.listeners[listener.GetListenerId()] = listener
	return subscriber.listeners[listener.GetListenerId()].Listen(output)
}

func (subscriber *Subscriber) Unsubscribe(key string) error {
	listener, exists := subscriber.listeners[key]
	if !exists {
		return fmt.Errorf("not subscribed to %s event", listener.GetListenerId())
	}
	delete(subscriber.listeners, key)
	return listener.Stop()
}

func (subscriber Subscriber) UnsubscribeAll() error {
	var err error
	err = nil
	for key := range subscriber.listeners {
		subErr := subscriber.Unsubscribe(key)
		if subErr != nil {
			err = subErr
		}
	}

	return err
}
