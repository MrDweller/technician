package event

type Listener interface {
	Listen(address string, port int, event Event, output chan<- []byte) error
	Stop() error
}
