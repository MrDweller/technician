package eventhandling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MrDweller/technician/workhandler"
	"github.com/gin-gonic/gin"
)

const USER_INTERACTIVE_EVENT_HANDLING EventHandlingSystemType = "USER_INTERACTIVE_EVENT_HANDLING"

// User interactive event handling system, notifies an external endpoint of the work task, and waits fo a response befoer assigning the work task.

type UserInteractiveEventHandling struct {
	WorkerId string
	workhandler.WorkHandler

	Address string
	Port    int

	DomainAddress string
	DomainPort    int

	ExternalEndpointUrl string
}

func (e *UserInteractiveEventHandling) InitEventHandler() error {
	router := gin.Default()
	router.POST("/work/take", e.takeWork)
	url := fmt.Sprintf("%s:%d", e.Address, e.Port)
	log.Printf("Starting event handling api on url: %s\n", url)

	go func() {
		err := router.Run(url)
		log.Printf("something wrong when running the event handling api: %s\n", err)

		time.Sleep(time.Second * 10)
		log.Printf("restarting the event handling api...\n")
		e.InitEventHandler()
	}()
	return nil
}

func (e *UserInteractiveEventHandling) HandleEvent(event WorkEvent) error {
	notifyExternalEndpointOfWorkTaskDTO := NotifyExternalEndpointOfWorkTaskDTO{
		WorkTaskType:        event.EventType,
		TecnicianSystemSlug: e.WorkerId,
		MowerSystemSlug:     event.ProductId,
		WorkTaskId:          event.WorkId,
		TakeWorkUrl:         fmt.Sprintf("http://%s:%d/work/take", e.DomainAddress, e.DomainPort),
	}

	rawBody, _ := json.Marshal(notifyExternalEndpointOfWorkTaskDTO)
	requestBody := bytes.NewBuffer(rawBody)

	response, err := http.Post(e.ExternalEndpointUrl, "application/json", requestBody)
	if err != nil {
		return err
	}

	log.Printf("response code from notify external enpoint of work task: %s\n", response.Status)
	log.Printf("response from notify external enpoint of work task: %s\n", response.Body)

	return nil
}

func (e *UserInteractiveEventHandling) takeWork(c *gin.Context) {
	var takeWorkDTO TakeWorkDTO
	if err := c.BindJSON(&takeWorkDTO); err != nil {
		c.Error(err)
		return
	}
	_, err := e.AssignWorker(takeWorkDTO.WorkId, e.WorkerId)
	if err != nil {
		log.Printf("Error during assignment of worker: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(200)
}

type TakeWorkDTO struct {
	WorkId string `json:"workId"`
}

type WorkTakenDTO struct {
	WorkId    string    `json:"workId"`
	ProductId string    `json:"productId"`
	EventType string    `json:"eventType"`
	Address   string    `json:"address"`
	StartTime time.Time `json:"startTime"`
}

type NotifyExternalEndpointOfWorkTaskDTO struct {
	WorkTaskType        string `json:"workTaskType"`
	TecnicianSystemSlug string `json:"tecnicianSystemSlug"`
	MowerSystemSlug     string `json:"mowerSystemSlug"`
	WorkTaskId          string `json:"workTaskId"`
	TakeWorkUrl         string `json:"takeWorkUrl"`
}
