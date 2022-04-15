package graph

import (
	"sync"
)

type SchedulerConfiguration struct {
	WorkChannelSize       uint16
	WorkResultChannelSize uint16
	GoroutinePoolSize     uint16
}

type WorkResult struct {
	Id  int64
	Err error
}

type WorkUnit func() error

type IScheduler interface {
	Schedule(id int64, work WorkUnit)

	Start() <-chan WorkResult

	// A blocking shutdown. Can be called only once.
	Shutdown()
}

type Scheduler struct {
	config         *SchedulerConfiguration
	workChan       chan func()
	workResultChan chan WorkResult

	poolWg sync.WaitGroup

	stopped bool
	stopMx  sync.RWMutex
}

func NewScheduler(config *SchedulerConfiguration) *Scheduler {
	if config == nil {
		config = &SchedulerConfiguration{}
	}

	if config.WorkChannelSize == 0 {
		config.WorkChannelSize = 10
	}

	if config.WorkResultChannelSize == 0 {
		config.WorkResultChannelSize = 10
	}

	if config.GoroutinePoolSize == 0 {
		config.GoroutinePoolSize = 10
	}

	return &Scheduler{
		config:         config,
		workChan:       make(chan func(), 10),
		workResultChan: make(chan WorkResult, 10),
	}
}

func (s *Scheduler) Schedule(id int64, work WorkUnit) {
	s.stopMx.RLock()
	defer s.stopMx.RUnlock()

	if s.stopped {
		// TODO: Maybe log an error.
		return
	}

	s.workChan <- func() {
		err := work()
		s.workResultChan <- WorkResult{
			Err: err,
			Id:  id,
		}
	}
}

func (s *Scheduler) Start() <-chan WorkResult {
	for i := 0; i < int(s.config.GoroutinePoolSize); i++ {
		s.poolWg.Add(1)

		go func() {
			defer s.poolWg.Done()

			for work := range s.workChan {
				work()
			}
		}()
	}

	return s.workResultChan
}

func (s *Scheduler) Shutdown() {
	s.stopMx.Lock()
	defer s.stopMx.Unlock()

	s.stopped = true
	close(s.workChan)
	s.poolWg.Wait()
	close(s.workResultChan)
}
