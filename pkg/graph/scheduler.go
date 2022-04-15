package graph

type WorkResult struct {
	Id  int64
	err error
}

type WorkUnit func() error

type IScheduler interface {
	Schedule(id int64, work WorkUnit)

	Start() chan WorkResult

	// A blocking shutdown.
	Shutdown()
}

type Scheduler struct {
}
