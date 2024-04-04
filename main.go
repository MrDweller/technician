package main

import (
	"io"
	"log"
	"os"
	"strconv"

	models "github.com/MrDweller/service-registry-connection/models"
	"github.com/joho/godotenv"

	"github.com/MrDweller/technician/cli"
	"github.com/MrDweller/technician/eventhandling"
	"github.com/MrDweller/technician/technician"
	"github.com/MrDweller/technician/workhandler"
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

	domainaddress := os.Getenv("DOMAIN_ADDRESS")
	domainPort, err := strconv.Atoi(os.Getenv("DOMAIN_PORT"))
	if err != nil {
		log.Panic(err)
	}
	systemName := os.Getenv("SYSTEM_NAME")

	serviceRegistryAddress := os.Getenv("SERVICE_REGISTRY_ADDRESS")
	serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		log.Panic(err)
	}

	eventHandlingSystemType := os.Getenv("EVENT_HANDLING_SYSTEM_TYPE")
	workHandlerType := os.Getenv("WORK_HANDLER_TYPE")

	var output io.Writer = os.Stdout
	technician, err := technician.NewTechnician(
		domainaddress,
		domainPort,
		systemName,
		serviceRegistryAddress,
		serviceRegistryPort,
		eventhandling.EventHandlingSystemType(eventHandlingSystemType),
		workhandler.WorkHandlerType(workHandlerType),
		output,
	)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		err := technician.StartTechnician()
		if err != nil {
			log.Panic(err)
		}
	}()

	cli.StartCli(technician, output)
}
