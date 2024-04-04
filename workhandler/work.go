package workhandler

import "time"

type Work struct {
	WorkId    string    `json:"workId"`
	ProductId string    `json:"productId"`
	EventType string    `json:"eventType"`
	Address   string    `json:"address"`
	StartTime time.Time `json:"startTime"`
}
