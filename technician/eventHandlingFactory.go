package technician

import (
	"fmt"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/technician/eventhandling"
	"github.com/MrDweller/technician/workhandler"
)

func NewEventHandlingSystem(eventHandlingSystemType eventhandling.EventHandlingSystemType, workHandlerType workhandler.WorkHandlerType, technician *Technician, certificateInfo orchestratormodels.CertificateInfo) (eventhandling.EventHandlingSystem, error) {
	workHandler, err := NewWorkHandler(workHandlerType, technician, certificateInfo)
	if err != nil {
		return nil, err
	}

	switch eventHandlingSystemType {
	case eventhandling.DIRECT_EVENT_HANDLING:
		return &eventhandling.DirectEventHandling{
			WorkerId:    technician.SystemName,
			WorkHandler: workHandler,
		}, nil
	default:
		return nil, fmt.Errorf("no implementation for the event handling system: %s", eventHandlingSystemType)
	}
}
