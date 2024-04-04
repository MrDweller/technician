package eventhandling

import (
	"net/http"
	"time"

	"github.com/MrDweller/technician/event"
	"github.com/MrDweller/technician/workhandler"
	"github.com/gin-gonic/gin"
)

const USER_INTERACTIVE_EVENT_HANDLING EventHandlingSystemType = "USER_INTERACTIVE_EVENT_HANDLING"

type UserInteractiveEventHandling struct {
	WorkerId string
	workhandler.WorkHandler
}

func (e *UserInteractiveEventHandling) HandleEvent(event event.Event) error {
	_, err := e.AssignWorker(event.WorkId, e.WorkerId)
	if err != nil {
		return err
	}

	return nil
}

func (e *UserInteractiveEventHandling) takeWork(c *gin.Context) {
	var takeWorkDTO TakeWorkDTO
	if err := c.BindJSON(&takeWorkDTO); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err := e.AssignWorker(takeWorkDTO.WorkId, e.WorkerId)
	if err != nil {
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
