package eventhandling

type EventHandlingSystemType string

type EventHandlingSystem interface {
	InitEventHandler() error
	HandleEvent(event WorkEvent) error
}

type WorkEvent struct {
	EventType string `json:"eventType"`
	WorkId    string `json:"workId"`
	ProductId string `json:"productId"`
}
