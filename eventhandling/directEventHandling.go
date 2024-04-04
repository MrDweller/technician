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

func (d *DirectEventHandling) HandleEvent(event event.Event) error {
	_, err := d.AssignWorker(event.WorkId, d.WorkerId)
	return err
}
