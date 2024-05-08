package pool

import (
	"fmt"
	"hago-plat/pcom/util"
	"sync/atomic"

	"git.ihago.top/ihago/ylog"
)

// CoPoolDispatchMode ...
type CoPoolDispatchMode int

// CbFunc ...
type CbFunc func(params ...interface{})

// CoPoolDispatchMode
const (
	CoPoolDispatchModeUID  = 1
	CoPoolDispatchModeRand = 2
)

// Task ...
type Task struct {
	Data []interface{}
	Cb   CbFunc
}

// CoPool ...
type CoPool struct {
	CoroutineNum uint32
	QueueLen     uint32
	Ch           []chan *Task
	Mode         CoPoolDispatchMode
	curidx       uint32
}

// worker ...
func (imp *CoPool) worker(idx int) {
	for task := range imp.Ch[idx] {
		func() {
			defer util.Panic()
			task.Cb(task.Data...)
		}()
	}
}

// Dispatch ...
func (imp *CoPool) Dispatch(uid int64, task *Task) {
	switch imp.Mode {
	case CoPoolDispatchModeUID:
		idx := uid % (int64)(imp.CoroutineNum)
		imp.Ch[idx] <- task
	case CoPoolDispatchModeRand:
		idx := imp.curidx % (uint32)(imp.CoroutineNum)
		imp.Ch[idx] <- task
		atomic.AddUint32(&imp.curidx, 1)
	default:
		ylog.Error("co_pool mode unknown")
	}
}

// RandDispatch ...
func (imp *CoPool) RandDispatch(task *Task) {
	idx := imp.curidx % (uint32)(imp.CoroutineNum)
	select {
	case imp.Ch[idx] <- task:
	default:
		ylog.Error(fmt.Sprintf("idx %v channel full", idx))
	}
	atomic.AddUint32(&imp.curidx, 1)
}

// NewCoPool ...
func NewCoPool(coroutineNum uint32, queueLen uint32, mode CoPoolDispatchMode) (pool *CoPool) {
	pool = &CoPool{
		CoroutineNum: coroutineNum,
		QueueLen:     queueLen,
		Ch:           make([]chan *Task, coroutineNum),
		Mode:         mode,
		curidx:       0,
	}

	var i uint32
	for i = 0; i < coroutineNum; i++ {
		pool.Ch[i] = make(chan *Task, queueLen)
		idx := (int)(i)
		go pool.worker(idx)
	}
	ylog.Info(fmt.Sprintf("CoPool startup coroutineNum:%d queueLen:%d mode:%d", coroutineNum, queueLen, mode))
	return pool
}
