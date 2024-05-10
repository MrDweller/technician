package technician

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	eventsubscriber "github.com/MrDweller/event-handler/subscriber"
	eventhandlertypes "github.com/MrDweller/event-handler/types"
	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"
	"github.com/MrDweller/technician/eventhandling"
	"github.com/MrDweller/technician/workhandler"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type Technician struct {
	eventhandling.EventHandlingSystem
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection
	output                    io.Writer

	SystemAddress string
	SystemPort    int

	eventSubscriber eventsubscriber.EventSubscriber
}

func NewTechnician(address string, port int, domainAddress string, domainPort int, systemName string, serviceRegistryAddress string, serviceRegistryPort int, eventHandlingSystemType eventhandling.EventHandlingSystemType, workHandlerType workhandler.WorkHandlerType, output io.Writer) (*Technician, error) {
	systemDefinition := models.SystemDefinition{
		Address:    domainAddress,
		Port:       domainPort,
		SystemName: systemName,
	}

	serviceRegistryConnection, err := serviceregistry.NewConnection(serviceregistry.ServiceRegistry{
		Address: serviceRegistryAddress,
		Port:    serviceRegistryPort,
	}, serviceregistry.SERVICE_REGISTRY_ARROWHEAD_4_6_1, models.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	serviceQueryResult, err := serviceRegistryConnection.Query(models.ServiceDefinition{
		ServiceDefinition: "orchestration-service",
	})
	if err != nil {
		return nil, err
	}

	serviceQueryData := serviceQueryResult.ServiceQueryData[0]

	orchestrationConnection, err := orchestrator.NewConnection(orchestrator.Orchestrator{
		Address: serviceQueryData.Provider.Address,
		Port:    serviceQueryData.Provider.Port,
	}, orchestrator.ORCHESTRATION_ARROWHEAD_4_6_1, orchestratormodels.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	eventSubscriber, err := eventsubscriber.EventSubscriberFactory(
		eventhandlertypes.EventHandlerImplementationType(os.Getenv("EVENT_HANDLER_IMPLEMENTATION")),
		domainAddress,
		domainPort,
		systemName,
		serviceRegistryAddress,
		serviceRegistryPort,
		serviceregistry.SERVICE_REGISTRY_ARROWHEAD_4_6_1,
		os.Getenv("CERT_FILE_PATH"),
		os.Getenv("KEY_FILE_PATH"),
		os.Getenv("TRUSTSTORE_FILE_PATH"),
	)
	if err != nil {
		return nil, err
	}

	technician := &Technician{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,
		EventHandlingSystem:       nil,
		output:                    output,

		SystemAddress: address,
		SystemPort:    port,

		eventSubscriber: eventSubscriber,
	}
	eventHandlingSystem, err := NewEventHandlingSystem(
		eventhandling.EventHandlingSystemType(eventHandlingSystemType),
		workHandlerType,
		technician,
		orchestratormodels.CertificateInfo{
			CertFilePath: os.Getenv("CERT_FILE_PATH"),
			KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
			Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
		},
	)
	if err != nil {
		return nil, err
	}

	technician.EventHandlingSystem = eventHandlingSystem

	return technician, nil
}

func (technician *Technician) ReceiveEvent(event []byte) {
	var workEventDto WorkDTO
	if err := json.Unmarshal(event, &workEventDto); err != nil {
		fmt.Fprintf(technician.output, "\n\t[!] Error received event with unkown structure: %s\n", event)
		return
	}

	fmt.Fprintf(technician.output, "\n\t[x] Received %s.\n", workEventDto)
	err := technician.EventHandlingSystem.HandleEvent(
		eventhandling.WorkEvent{
			EventType: workEventDto.EventType,
			WorkId:    workEventDto.WorkId,
			ProductId: workEventDto.ProductId,
		},
	)
	if err != nil {
		fmt.Fprintf(technician.output, "\n\t[!] Error during handling of the event: %s\n", err)

	}
}

func (technician *Technician) StopTechnician() error {
	return technician.eventSubscriber.UnregisterEventSubscriberSystem()
}

func (technician *Technician) Subscribe(requestedService string) error {
	fmt.Fprintf(technician.output, "\n\t[*] Subscribing to %s events.\n", requestedService)
	return technician.eventSubscriber.Subscribe(eventhandlertypes.EventType(requestedService), technician)

}

func (technician *Technician) Unsubscribe(requestedService string) error {
	err := technician.eventSubscriber.Unsubscribe(eventhandlertypes.EventType(requestedService))
	if err != nil {
		return err
	}
	fmt.Fprintf(technician.output, "\n\t[*] Unsubscribing from %s events.\n", requestedService)
	return nil
}

type WorkDTO struct {
	WorkId string `json:"workId"`

	ProductId string `json:"productId"`
	EventType string `json:"eventType"`
}
