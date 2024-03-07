package main

import (
	"io"
	"log"
	"os"
	"strconv"

	models "github.com/MrDweller/service-registry-connection/models"
	"github.com/joho/godotenv"

	"github.com/MrDweller/technician/cli"
	"github.com/MrDweller/technician/technician"
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

	address := os.Getenv("ADDRESS")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panic(err)
	}
	systemName := os.Getenv("SYSTEM_NAME")

	serviceRegistryAddress := os.Getenv("SERVICE_REGISTRY_ADDRESS")
	serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		log.Panic(err)
	}

	var output io.Writer = os.Stdout
	technician, err := technician.NewTechnician(address, port, systemName, serviceRegistryAddress, serviceRegistryPort, output)
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
