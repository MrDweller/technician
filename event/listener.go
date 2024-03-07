package event

type Listener interface {
	Listen(address string, port int, event Event, metadata map[string]string, output chan<- []byte) error
	Stop() error
}
