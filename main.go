package workerpool

import "context"

type Task interface {
	Execute()
}

type Service struct {
	dispacher dispatcher
	async     bool
	Exit      context.CancelFunc
}

// Start the dispacher and make de executor available to receive tasks.
func (s *Service) Run() {
	if s.async {
		go s.dispacher.asyncDispatch()
	} else {
		go s.dispacher.dispatch()
	}
}

// Send one task to the dispacher.
// This block if the queue depth reach his limit and option "safe=true".
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

	return &Service{dispacher: dispatcher, async: false, Exit: cancel}
}

// Make dispatcher run async.
// This mean the underliying go routine will block until a worker release the queue.
func (s *Service) WithAsyncDispatch() *Service {
	s.async = true
	return s
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

func (d *dispatcher) asyncDispatch() {
	for {
		select {
		case task := <-d.taskQueue:
			go func(task Task) {
				workQueue := <-d.workPool

				workQueue <- task

			}(task)
		case <-d.ctx.Done():
			return
		}
	}
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
