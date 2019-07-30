package jobqueue

import (
	"sync"
	"time"
)

type workerSig struct{}
type workerSocket struct {
	terminate chan workerSig
	suspend   chan workerSig
	resume    chan workerSig
}

// queue in an internal Queue implementation
type queue struct {
	jobCapacity     uint
	workersCapacity uint

	jobTimeout       time.Duration
	retryDelay       time.Duration
	requeueIfTimeout bool

	succeedHandler SucceedHandler
	dropHandler    DropHandler
	errHandler     ErrHandler
	panicHandler   PanicHandler

	rootWorkers workers
	sync        sync.RWMutex

	jobq    chan Job
	workerq chan workerSocket

	suspended bool
	closed    bool
}

// Sync return a channel synchronized with the job queue.
func (q *queue) Sync() chan<- Job {
	q.sync.RLock()
	defer q.sync.RUnlock()

	sync := make(chan Job)
	if q.closed {
		// if jobq is closed, ignore new job
		go func() { defer close(sync); <-sync }()
	} else {
		// TODO(ani): recover to avoid panic
		go func() { defer close(sync); q.jobq <- <-sync }()
	}
	return sync
}

// Async return a temporary channel with timeout, synchronized
// with the job queue.
func (q *queue) Async(timeout time.Duration) chan<- Job {
	q.sync.RLock()
	defer q.sync.RUnlock()

	async := make(chan Job)
	if q.closed {
		// if jobq is closed, ignore new job
		go func() { defer close(async); <-async }()
	} else {
		// TODO(ani): recover to avoid panic
		go func() {
			defer close(async)
			select {
			case q.jobq <- <-async:
			case <-time.After(timeout):
			}
		}()
	}
	return async
}

// Scale adds or removes worker to reach the given value.
func (q *queue) Scale(workers uint) (int, error) {
	q.sync.Lock()
	defer q.sync.Unlock()

	delta := int(workers) - q.Workers()
	if delta > 0 {
		return q.addWorkers(delta)
	}
	return q.removeWorker(-delta)
}

// Close flushes and closes the job queue and stop all workers.
func (q *queue) Close() error {
	q.sync.Lock()
	q.closed = true

	close(q.jobq)
	for range q.jobq {
	}

	q.sync.Unlock()
	_, _ = q.Scale(0)
	return nil
}

// WaitAndClose waits the job queue to be empty before closing all workers.
func (q *queue) WaitAndClose() error {
	q.sync.Lock()
	q.closed = true
	close(q.jobq)
	q.sync.Unlock()

	for q.JobLoad() > 0 {
		time.Sleep(50 * time.Millisecond)
	}

	_, _ = q.Scale(0)
	return nil
}

// SuspendWorkers suspend all workers.
func (q *queue) SuspendWorkers() {
	q.sync.Lock()
	defer q.sync.Unlock()

	// ignore if already suspended
	if q.suspended {
		return
	}

	workers := len(q.workerq)
	for i := 0; i < workers; i++ {
		worker := <-q.workerq
		worker.suspend <- workerSig{}
		q.workerq <- worker
	}
	q.suspended = true
}

// ResumeWorkers resume all workers.
func (q *queue) ResumeWorkers() {
	q.sync.Lock()
	defer q.sync.Unlock()

	// ignore if not suspended
	if !q.suspended {
		return
	}

	workers := len(q.workerq)
	for i := 0; i < workers; i++ {
		worker := <-q.workerq
		worker.resume <- workerSig{}
		q.workerq <- worker
	}
	q.suspended = true
}

// Workers returns the number of worker in worker queue.
func (q *queue) Workers() int { return len(q.workerq) }

// JobCapacity returns the number of maximum jobs in job queue.
func (q *queue) JobCapacity() int { return cap(q.jobq) }

// JobLoad returns the number of jobs in job queue.
func (q *queue) JobLoad() int { return len(q.jobq) }

// addWorkers add N workers to the worker queue.
func (q *queue) addWorkers(n int) (int, error) {
	for i := 0; i < n; i++ {
		if len(q.workerq) == n {
			return i, newErrMaxWorkerReached()
		}

		workers := q.rootWorkers.copy()
		if err := workers.initialize(); err != nil {
			return i, err
		}

		wkch := workerSocket{
			terminate: make(chan workerSig, 1),
			suspend:   make(chan workerSig, 1),
			resume:    make(chan workerSig, 1),
		}
		go q.do(workers, wkch)
		q.workerq <- wkch
	}

	return n, nil
}

// removeWorker remove N workers from the worker queue.
func (q *queue) removeWorker(workers int) (int, error) {
	for i := 0; i < workers; i++ {
		if len(q.workerq) == 0 {
			return i, newErrMinWorkerReached()
		}

		worker := <-q.workerq
		worker.terminate <- workerSig{}
		close(worker.terminate)
		close(worker.suspend)
		close(worker.resume)
	}

	return workers, nil
}

// workers simplify processes with several workers
type workers []Worker

// copy make a copy of all workers
func (ws workers) copy() workers {
	copy := make([]Worker, len(ws))

	for i, w := range ws {
		copy[i] = w.Copy()
	}
	return copy
}

// initialize all workers
func (ws workers) initialize() error {
	for _, w := range ws {
		if err := w.Initialize(); err != nil {
			return err
		}
	}
	return nil
}

// terminate all workers
func (ws workers) terminate() {
	for _, w := range ws {
		w.Terminate()
	}
}
