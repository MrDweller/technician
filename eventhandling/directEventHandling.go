package eventhandling

import (
	"github.com/MrDweller/technician/event"
	"github.com/MrDweller/technician/workhandler"
)

const DIRECT_EVENT_HANDLING EventHandlingSystemType = "DIRECT_EVENT_HANDLING"

type DirectEventHandling struct {
	WorkerId string
	workhandler.ExternalWorkHandler
}

func (d *DirectEventHandling) HandleEvent(event event.Event) error {
	return d.AssignWorker(event.WorkId, d.WorkerId)
}
