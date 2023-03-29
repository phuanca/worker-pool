package workerpool

import "context"

type Task interface {
	Execute()
}

type Service struct {
	dispacher dispatcher
	exit      context.CancelFunc
}

// Start the dispacher and make de executor available to receive tasks.
func (s *Service) Run() {
	go s.dispacher.dispatch()

}

// Stop all workers and exit.
func (s *Service) Exit() {
	s.exit()
}

// Send one task to the dispacher.
// This block if the queue depth reach his limit.
func (s *Service) Send(task Task) {
	s.dispacher.taskQueue <- task
}

// Create an executor service with the given maxQueueDepth and the maxWorkers values.
func NewService(maxWorkers, maxQueueDepth int) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	taskQueue := make(chan Task, maxQueueDepth)
	workPool := make(chan chan Task, maxWorkers)
	dispatcher := dispatcher{ctx: ctx, taskQueue: taskQueue, workPool: workPool}

	for i := 0; i < maxWorkers; i++ {
		w := worker{
			ctx:       ctx,
			number:    i,
			workQueue: make(chan Task),
			workPool:  workPool,
		}
		go w.start()
	}

	return &Service{dispacher: dispatcher, exit: cancel}
}

type worker struct {
	ctx       context.Context
	number    int
	workQueue chan Task
	workPool  chan chan Task
}

func (w *worker) start() {
	for {
		// notify ready to work.
		w.workPool <- w.workQueue
		select {
		case workToDo := <-w.workQueue:
			workToDo.Execute()
		case <-w.ctx.Done():
			return
		}
	}
}

type dispatcher struct {
	ctx       context.Context
	taskQueue chan Task
	workPool  chan chan Task
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case task := <-d.taskQueue:
			workQueue := <-d.workPool
			workQueue <- task
		case <-d.ctx.Done():
			return
		}
	}
}
