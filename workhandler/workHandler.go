package workhandler

type WorkHandlerType string

type WorkHandler interface {
	AssignWorker(workId string, workerId string) (*Work, error)
}
