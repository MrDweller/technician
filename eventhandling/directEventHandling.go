package eventhandling

import (
	"github.com/MrDweller/technician/event"
	"github.com/MrDweller/technician/workhandler"
)

const DIRECT_EVENT_HANDLING EventHandlingSystemType = "DIRECT_EVENT_HANDLING"

type DirectEventHandling struct {
	WorkerId string
	workhandler.WorkHandler
}

func (e *DirectEventHandling) InitEventHandler() error {
	return nil
}

func (e *DirectEventHandling) HandleEvent(event event.Event) error {
	_, err := e.AssignWorker(event.WorkId, e.WorkerId)
	return err
}
