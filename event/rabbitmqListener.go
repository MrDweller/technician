package event

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqListener struct {
	done chan bool
}

func NewRabbitmqListener() *RabbitmqListener {
	done := make(chan bool)
	return &RabbitmqListener{
		done: done,
	}
}

func (listener *RabbitmqListener) Listen(address string, port int, event Event, output chan<- []byte) error {
	url := fmt.Sprintf("%s:%d/", address, port)
	dialAddrr := fmt.Sprintf("amqp://guest:guest@%s", url)
	conn, err := amqp.Dial(dialAddrr)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			// log.Println("TESTY1")

			fmt.Printf(" [x] received %s, from %s event.\n", d.Body, event.Name)
			// output <- d.Body
		}
	}()

	fmt.Printf(" [*] Subscribed to %s events, listening for updates..\n", event.Name)

	<-listener.done
	return nil

}

func (listener *RabbitmqListener) Stop() error {
	listener.done <- true
	return nil
}
