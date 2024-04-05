package technician

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"
	"github.com/MrDweller/technician/event"
	"github.com/MrDweller/technician/eventhandling"
	"github.com/MrDweller/technician/workhandler"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type Technician struct {
	eventhandling.EventHandlingSystem
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection
	*event.Subscriber
	eventChannel chan []byte
	output       io.Writer

	SystemAddress string
	SystemPort    int
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

	technician := &Technician{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,
		Subscriber:                event.NewSubscriber(),
		eventChannel:              make(chan []byte),
		EventHandlingSystem:       nil,
		output:                    output,

		SystemAddress: address,
		SystemPort:    port,
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

func (technician *Technician) StartTechnician() error {
	_, err := technician.ServiceRegistryConnection.RegisterSystem(technician.SystemDefinition)
	if err != nil {
		return err
	}

	for receivedEvent := range technician.eventChannel {
		var event event.Event
		if err := json.Unmarshal(receivedEvent, &event); err != nil {
			fmt.Fprintf(technician.output, "\n\t[!] Error received event with unkown structure: %s\n", receivedEvent)
			continue
		}

		fmt.Fprintf(technician.output, "\n\t[x] Received %s.\n", event)
		err := technician.EventHandlingSystem.HandleEvent(event)
		if err != nil {
			fmt.Fprintf(technician.output, "\n\t[!] Error during handling of the event: %s\n", err)

		}
	}

	return nil
}

func (technician *Technician) StopTechnician() error {
	err := technician.Subscriber.UnsubscribeAll()
	if err != nil {
		return err
	}

	err = technician.ServiceRegistryConnection.UnRegisterSystem(technician.SystemDefinition)
	if err != nil {
		return err
	}

	return err
}

func (technician *Technician) Subscribe(requestedService string) error {
	orchestrationResponse, err := technician.OrchestrationConnection.Orchestration(
		requestedService,
		[]string{
			"AMQP-INSECURE-JSON",
		},
		orchestratormodels.SystemDefinition{
			Address:    technician.Address,
			Port:       technician.Port,
			SystemName: technician.SystemName,
		},
		orchestratormodels.AdditionalParametersArrowhead_4_6_1{
			OrchestrationFlags: map[string]bool{
				"overrideStore": true,
			},
		},
	)
	if err != nil {
		return err
	}

	if len(orchestrationResponse.Response) <= 0 {
		return errors.New("found no providers")
	}
	providers := orchestrationResponse.Response

	for _, provider := range providers {
		fmt.Fprintf(technician.output, "\n\t[*] Subscribing to %s events on %s at %s:%d.\n", requestedService, provider.Provider.SystemName, provider.Provider.Address, provider.Provider.Port)
		go func(systemName string, address string, port int, serviceDefinition string, metadata map[string]string) {
			err := technician.Subscriber.Subscribe(
				systemName,
				address,
				port,
				event.EventDefinition{
					EventType: serviceDefinition,
				},
				metadata,
				technician.eventChannel,
			)

			if err != nil {
				fmt.Fprintf(technician.output, "\n\t[*] Error during subscription: %s\n", err)
				return
			}

		}(provider.Provider.SystemName, provider.Provider.Address, provider.Provider.Port, requestedService, provider.Metadata)

	}

	return nil

}

func (technician *Technician) Unsubscribe(requestedService string) error {
	err := technician.Subscriber.UnsubscribeAllByEvent(
		event.EventDefinition{
			EventType: requestedService,
		},
	)
	if err != nil {
		return err
	}
	fmt.Fprintf(technician.output, "\n\t[*] Unsubscribing from %s events.\n", requestedService)
	return nil
}
