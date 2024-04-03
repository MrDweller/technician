package workhandler

type WorkHandler interface {
	AssignWorker(workId string, workerId string) error
}
