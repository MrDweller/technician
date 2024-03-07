package technician

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"
	"time"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"
	"github.com/MrDweller/technician/event"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type Technician struct {
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection
	*event.Subscriber
	eventChannel chan []byte
}

func NewTechnician(address string, port int, systemName string, serviceRegistryAddress string, serviceRegistryPort int) (*Technician, error) {
	systemDefinition := models.SystemDefinition{
		Address:    address,
		Port:       port,
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

	return &Technician{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,
		Subscriber:                event.NewSubscriber(),
	}, nil
}

func (technician *Technician) StartTechnician() error {
	_, err := technician.ServiceRegistryConnection.RegisterSystem(technician.SystemDefinition)
	if err != nil {
		return err
	}

	// for event := range technician.eventChannel {
	// 	fmt.Printf(" [x] received %s.\n", event)
	// }
	return nil
}

func (technician *Technician) StopTechnician() error {
	err := technician.ServiceRegistryConnection.UnRegisterSystem(technician.SystemDefinition)
	if err != nil {
		return err
	}
	return nil
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
	provider := orchestrationResponse.Response[0]

	err = technician.Subscriber.Subscribe(
		provider.Provider.Address,
		provider.Provider.Port,
		event.Event{
			Name: requestedService,
		},
		technician.eventChannel,
	)
	if err != nil {
		return err
	}

	return nil

}

func (technician *Technician) Unsubscribe(requestedService string) error {
	return technician.Subscriber.Unsubscribe(event.Event{
		Name: requestedService,
	})
}

func (technician *Technician) getClient() (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(os.Getenv("CERT_FILE_PATH"), os.Getenv("KEY_FILE_PATH"))
	if err != nil {
		return nil, err
	}

	// Load truststore.p12
	truststoreData, err := os.ReadFile(os.Getenv("TRUSTSTORE_FILE_PATH"))
	if err != nil {
		return nil, err

	}

	// Extract the root certificate(s) from the truststore
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(truststoreData); !ok {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            pool,
				InsecureSkipVerify: false,
			},
		},
	}
	return client, nil
}
