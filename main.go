package main

import (
	"log"
	"os"
	"strconv"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	models "github.com/MrDweller/service-registry-connection/models"
	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
	"github.com/joho/godotenv"

	event "github.com/MrDweller/technician/event"
)

type EventData struct {
	EventType string                  `json:"eventType"`
	Payload   string                  `json:"payload"`
	Source    models.SystemDefinition `json:"source"`
	TimeStamp string                  `json:"timeStamp"`
}

type SubscriberData struct {
	EventType        string                  `json:"eventType"`
	MatchMetaData    bool                    `json:"matchMetaData"`
	NotifyUri        string                  `json:"notifyUri"`
	SubscriberSystem models.SystemDefinition `json:"subscriberSystem"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	serviceRegistryAddress := os.Getenv("SERVICE_REGISTRY_ADDRESS")
	serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		log.Panic(err)
	}

	serviceRegistryConnection, err := serviceregistry.NewConnection(
		serviceregistry.ServiceRegistry{
			Address: serviceRegistryAddress,
			Port:    serviceRegistryPort,
		},
		serviceregistry.ServiceRegistryImplementationType(os.Getenv("SERVICE_REGISTRY_IMPLEMENTATION")),
		models.CertificateInfo{
			CertFilePath: os.Getenv("CERT_FILE_PATH"),
			KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
			Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
		},
	)
	if err != nil {
		log.Panic(err)
	}

	systemPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panic(err)
	}

	system := models.SystemDefinition{
		Address:            os.Getenv("ADDRESS"),
		Port:               systemPort,
		SystemName:         os.Getenv("SYSTEM_NAME"),
		AuthenticationInfo: "",
	}

	// response, err := serviceRegistryConnection.RegisterSystem(system)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// log.Printf("Registered system: %s\n", string(response[:]))

	serviceQueryResult, err := serviceRegistryConnection.Query(models.ServiceDefinition{
		ServiceDefinition: "orchestration-service",
	})
	if len(serviceQueryResult.ServiceQueryData) < 1 {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panicf("Found no orchestrator\n")
	}

	var orchestratorAddress string
	var orchestratorPort int
	for _, queryResult := range serviceQueryResult.ServiceQueryData {
		if queryResult.Provider.SystemName == "orchestrator" {
			orchestratorAddress = queryResult.Provider.Address
			orchestratorPort = queryResult.Provider.Port
			break
		}
	}

	if err != nil {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panic(err)
	}
	orchestratorConnection, err := orchestrator.NewConnection(
		orchestrator.Orchestrator{
			Address: orchestratorAddress,
			Port:    orchestratorPort,
		},
		orchestrator.OrchestratorImplementationType(os.Getenv("ORCHESTRATOR_IMPLEMENTATION")),
		orchestratormodels.CertificateInfo{
			CertFilePath: os.Getenv("CERT_FILE_PATH"),
			KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
			Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
		},
	)
	if err != nil {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panic(err)
	}

	orchestrationResponse, err := orchestratorConnection.Orchestration(
		"STUCK",
		[]string{
			"AMQP-INSECURE-JSON",
		},
		orchestratormodels.SystemDefinition{
			Address:    system.Address,
			Port:       system.Port,
			SystemName: system.SystemName,
		},
		orchestratormodels.AdditionalParametersArrowhead_4_6_1{
			OrchestrationFlags: map[string]bool{
				"overrideStore": true,
			},
		},
	)
	if err != nil {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panic(err)
	}
	log.Printf("Orchestration: %v\n", orchestrationResponse.Response)

	if len(orchestrationResponse.Response) <= 0 {
		log.Printf("Found no providers\n")
	}
	provider := orchestrationResponse.Response[0]

	event.Receive(provider.Provider.Address, provider.Provider.Port)
}
