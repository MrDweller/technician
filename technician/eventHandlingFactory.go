package technician

import (
	"fmt"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/technician/eventhandling"
	"github.com/MrDweller/technician/workhandler"
)

func NewEventHandlingSystem(eventHandlingSystemType eventhandling.EventHandlingSystemType, technician *Technician, certificateInfo orchestratormodels.CertificateInfo) (eventhandling.EventHandlingSystem, error) {
	switch eventHandlingSystemType {
	case eventhandling.DIRECT_EVENT_HANDLING:
		return &eventhandling.DirectEventHandling{
			WorkerId: technician.SystemName,
			ExternalWorkHandler: workhandler.ExternalWorkHandler{
				TakeWorkServiceDefinition: "assign-worker",
				OrchestrationConnection:   technician.OrchestrationConnection,
				SystemDefinition: orchestratormodels.SystemDefinition{
					Address:    technician.Address,
					Port:       technician.Port,
					SystemName: technician.SystemName,
				},
				CertificateInfo: certificateInfo,
			},
		}, nil
	default:
		return nil, fmt.Errorf("no implementation for the event handling system: %s", eventHandlingSystemType)
	}
}
