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

	system := models.SystemDefinition{
		Address:            "eventhandler-test-system",
		Port:               8080,
		SystemName:         "eventhandler-test-system",
		AuthenticationInfo: "",
	}

	response, err := serviceRegistryConnection.RegisterSystem(system)
	if err != nil {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panic(err)
	}
	log.Printf("Registered system: %s\n", string(response[:]))

	orchestratorAddress := os.Getenv("ORCHESTRATOR_ADDRESS")
	orchestratorPort, err := strconv.Atoi(os.Getenv("ORCHESTRATOR_PORT"))
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
		orchestratormodels.ServiceDefinition{
			ServiceDefinition: "",
		},
		orchestratormodels.SystemDefinition{
			Address:    system.Address,
			Port:       system.Port,
			SystemName: system.SystemName,
		},
		map[string]bool{
			"overrideStore":    true,
			"enableInterCloud": false,
		},
		orchestratormodels.RequesterCloud{},
	)
	if err != nil {
		serviceRegistryConnection.UnRegisterSystem(system)
		log.Panic(err)
	}
	log.Printf("Orchestration: %v\n", orchestrationResponse.Response)

	if len(orchestrationResponse.Response) <= 0 {
		log.Panicf("Found no providers\n")
	}
	provider := orchestrationResponse.Response[0]

	Receive(provider.Provider.Address, provider.Provider.Port)
}
