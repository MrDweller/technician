package eventhandling

import (
	"github.com/MrDweller/technician/event"
)

type EventHandlingSystemType string

type EventHandlingSystem interface {
	InitEventHandler() error
	HandleEvent(event event.Event) error
}
