package eventhandling

import (
	"github.com/MrDweller/technician/event"
)

type EventHandlingSystemType string

type EventHandlingSystem interface {
	HandleEvent(event event.Event) error
}
