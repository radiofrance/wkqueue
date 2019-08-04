package jobqueue

import (
	"runtime"
	"time"
)

// TODO(ani): add documentation
type Options interface {
	apply(*queue) error
}

var defaultsOpts = []Options{
	SetJobCapacity(16),
	SetWorkerCapacity(uint(runtime.NumCPU())),
	SetJobTimeout(time.Second),
	SetRetryDelay(time.Second),
	SetSuccessHandler(func(Worker, Job) {}),
	SetDropHandler(func(Worker, Job) {}),
	SetErrHandler(func(Worker, error, Job) {}),
	SetPanicHandler(func(Worker, interface{}, Job) {}),
}

type optionFunc func(*queue) error
type optionFuncNoErr func(*queue)

func (f optionFunc) apply(q *queue) error      { return f(q) }
func (f optionFuncNoErr) apply(q *queue) error { f(q); return nil }

func AddWorker(worker Worker) Options {
	return optionFunc(func(q *queue) error {
		if worker == nil {
			return newErrNilWorker()
		}
		q.rootWorkers = append(q.rootWorkers, worker)
		return nil
	})
}
func SetJobCapacity(capacity uint) Options {
	// don't need to close the channel (https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ)
	// because channel is override each time, all non closed chan will be garbage collected
	return optionFuncNoErr(func(q *queue) { q.jobq = make(chan Job, capacity) })
}
func SetWorkerCapacity(capacity uint) Options {
	// same as above
	return optionFuncNoErr(func(q *queue) { q.workerq = make(chan workerSocket, capacity) })
}
func SetJobTimeout(timeout time.Duration) Options {
	return optionFuncNoErr(func(q *queue) { q.jobTimeout = timeout })
}
func SetRetryDelay(delay time.Duration) Options {
	return optionFuncNoErr(func(q *queue) { q.retryDelay = delay })
}
func RequeueIfTimeout() Options {
	return optionFuncNoErr(func(q *queue) { q.requeueIfTimeout = true })
}
func SetSuccessHandler(handler SuccessHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("SuccessHandler")
		}
		q.succeedHandler = handler
		return nil
	})
}
func SetDropHandler(handler DropHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("DropHandler")
		}
		q.dropHandler = handler
		return nil
	})
}
func SetErrHandler(handler ErrHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("ErrHandler")
		}
		q.errHandler = handler
		return nil
	})
}
func SetPanicHandler(handler PanicHandler) Options {
	return optionFunc(func(q *queue) error {
		if handler == nil {
			return newErrNilHandler("PanicHandler")
		}
		q.panicHandler = handler
		return nil
	})
}
