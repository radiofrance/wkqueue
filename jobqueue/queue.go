// Package jobqueue is an implementation of a job centric working queue.
package jobqueue

import (
	"sync"
)

// SuccessHandler is called when a job is successfully consumed.
type SuccessHandler func(worker Worker, job *Job)

// ErrHandler is called when an error occurs during the job
// consumption.
type ErrHandler func(worker Worker, err error, job *Job)

// DropHandler is called job a retry too many times and will be
// not requeued in the job queue.
type DropHandler func(worker Worker, job *Job)

// PanicHandler is called a panic occurs.
type PanicHandler func(worker Worker, recover interface{}, job *Job)

// Queue is a container around the job queue; we can
// increase or decrease the number of workers, suspend
// them or send a job to the job queue.
type Queue interface {
	Sync() chan<- *Job

	Scale(uint) (int, error)

	SuspendWorkers()
	ResumeWorkers()

	Close()
	WaitAndClose()

	WorkersLimit() int
	NumWorkers() int
	JobCapacity() int
	JobLoad() int
}

// New returns a new job queue.
func New(opts ...Options) (Queue, error) {
	q := &queue{sync: sync.RWMutex{}}

	opts = append(defaultsOpts, opts...)
	for _, opt := range opts {
		if err := opt.apply(q); err != nil {
			return nil, err
		}
	}

	if len(q.rootWorkers) == 0 {
		return nil, newErrNoWorker()
	}
	return q, nil
}
