package iamqp

// Task 定义了协程池运行的任务接口
type Task interface {
	Launch()
}

type worker struct {
	WorkerPool  chan chan Task
	TaskChannel chan Task
	quit        chan bool
}

func newWorker(workerPool chan chan Task) worker {
	return worker{
		WorkerPool:  workerPool,
		TaskChannel: make(chan Task),
		quit:        make(chan bool),
	}
}

func (w *worker) start() {
	go func() {
		for {
			w.WorkerPool <- w.TaskChannel
			select {
			case job := <-w.TaskChannel:
				job.Launch()
			case <-w.quit:
				return
			}
		}
	}()
}

//func (w *worker) stop() {
//	go func() {
//		w.quit <- true
//	}()
//}

// Manager 协程池管理
type Manager struct {
	TaskQueue  chan Task
	WorkerPool chan chan Task
	maxWorkers int
}

// newManager return goroutine pool manager
func newManager(taskQueueSize, maxWorkers int) *Manager {
	jobQueue := make(chan Task, taskQueueSize)
	pool := make(chan chan Task, maxWorkers)
	return &Manager{TaskQueue: jobQueue, WorkerPool: pool, maxWorkers: maxWorkers}
}

// Run starts goroutine pool running
func (mgr *Manager) run() {
	for i := 0; i < mgr.maxWorkers; i++ {
		w := newWorker(mgr.WorkerPool)
		w.start()
	}
	go mgr.dispatch()
}

func (mgr *Manager) dispatch() {
	for task := range mgr.TaskQueue {
		go func(task Task) {
			worker := <-mgr.WorkerPool
			worker <- task
		}(task)
	}
}

// Dispatch a goroutine task
func (mgr *Manager) Dispatch(task Task) {
	mgr.TaskQueue <- task
}
