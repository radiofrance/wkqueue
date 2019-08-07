package jobqueue

import (
	"runtime"
	"sync"
	"time"
)

// Options configures the job queue.
type Options interface {
	apply(*queue) error
}

var defaultsOpts = []Options{
	SetJobCapacity(16),
	SetWorkerCapacity(uint(runtime.NumCPU())),
	SetJobTimeout(time.Second),
	SetRetryDelay(time.Second),
	SetSuccessHandler(func(Worker, *Job) {}),
	SetDropHandler(func(Worker, *Job) {}),
	SetErrHandler(func(Worker, error, *Job) {}),
	SetPanicHandler(func(Worker, interface{}, *Job) {}),
}

type optionFunc func(*queue) error
type optionFuncNoErr func(*queue)

func (f optionFunc) apply(q *queue) error      { return f(q) }
func (f optionFuncNoErr) apply(q *queue) error { f(q); return nil }

// AddWorker add the given worker to the work queue.
func AddWorker(worker Worker) Options {
	return optionFunc(func(q *queue) error {
		if worker == nil {
			return newErrNilWorker()
		}
		q.rootWorkers = append(q.rootWorkers, worker)
		return nil
	})
}

// SetJobCapacity sets the maximum number of job is the job queue.
func SetJobCapacity(capacity uint) Options {
	// don't need to close the channel (https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ)
	// because channel is override each time, all non closed chan will be garbage collected
	return optionFuncNoErr(func(q *queue) { q.jobq = make(chan *Job, capacity) })
}

// SetWorkerCapacity sets the maximum number of parallel workers.
func SetWorkerCapacity(capacity uint) Options {
	// same as above
	return optionFuncNoErr(func(q *queue) { q.workerq = make(chan workerSocket, capacity) })
}

// SetJobTimeout sets when a job timed out.
func SetJobTimeout(timeout time.Duration) Options {
	return optionFuncNoErr(func(q *queue) {
		q.jobTimeout = timeout
		q.timerPool = sync.Pool{New: func() interface{} { return time.NewTimer(time.Second) }}
	})
}

// SetRetryDelay sets the delaying duration before a job must be sent to the queue.
func SetRetryDelay(delay time.Duration) Options {
	return optionFuncNoErr(func(q *queue) { q.retryDelay = delay })
}

// RequeueIfTimeout tries to rerun a job which has timed out.
func RequeueIfTimeout() Options {
	return optionFuncNoErr(func(q *queue) { q.requeueIfTimeout = true })
}

// SetSuccessHandler sets the function which will be called when a job is successfully consumed.
func SetSuccessHandler(handler SuccessHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("SuccessHandler")
		}
		q.succeedHandler = handler
		return nil
	})
}

// SetDropHandler sets the function which will be called when a job is dropped.
func SetDropHandler(handler DropHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("DropHandler")
		}
		q.dropHandler = handler
		return nil
	})
}

// SetErrHandler sets the function which will be called when a worker returns an error or times out.
func SetErrHandler(handler ErrHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("ErrHandler")
		}
		q.errHandler = handler
		return nil
	})
}

// SetPanicHandler sets the function which will be called when a worker panics.
func SetPanicHandler(handler PanicHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("PanicHandler")
		}
		q.panicHandler = handler
		return nil
	})
}
