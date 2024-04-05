package technician

import (
	"fmt"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/technician/workhandler"
)

func NewWorkHandler(workHandlerType workhandler.WorkHandlerType, technician *Technician, certificateInfo orchestratormodels.CertificateInfo) (workhandler.WorkHandler, error) {
	switch workHandlerType {
	case workhandler.EXTERNAL_WORK_HANDLER:
		return &workhandler.ExternalWorkHandler{
			TakeWorkServiceDefinition: "assign-worker",
			OrchestrationConnection:   technician.OrchestrationConnection,
			SystemDefinition: orchestratormodels.SystemDefinition{
				Address:    technician.Address,
				Port:       technician.Port,
				SystemName: technician.SystemName,
			},
			CertificateInfo: certificateInfo,
		}, nil
	default:
		return nil, fmt.Errorf("no implementation for the work handler system: %s", workHandlerType)
	}
}
