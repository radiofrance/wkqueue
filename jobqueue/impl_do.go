package jobqueue

import (
	"time"
)

// asyncExec runs the job asynchronously
func (q *queue) asyncExec(wk Worker, job Job) error {
	cerr := make(chan error)

	// because worker must no be stucked or crashed, we need to run worker.Do
	// in a goroutine with panic recovering.
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				q.panicHandler(wk, r, job)
			}
		}()
		defer close(cerr)
		cerr <- wk.Do(job)
	}()

	// because job is a go routine, we need to wait a response (or timeout)
	select {
	case err, ok := <-cerr:
		if ok {
			return err
		}
		// chan returns ok == false only if it's empty and closed,
		// which occurs only if job panics.
		return newErrJobPanic()

	case <-time.After(q.jobTimeout):
		return newErrJobTimeout()
	}
}

// do is the main loop of the internal worker instance
func (q *queue) do(wks workers, wkch workerSocket) {
	defer wks.terminate()

	for {
		select {
		case job, ok := <-q.jobq:
			if !ok || job.isZero() {
				// ignore if the chan is closed (occurs only on termination) or if the job is empty
				continue
			}

			consumed := false
			for _, wk := range wks {
				if !wk.CanConsume(job) {
					continue
				}
				consumed = true

				err := q.asyncExec(wk, job)
				if err == nil {
					q.succeedHandler(wk, job)
					continue
				}

				q.errHandler(wk, err, job)
				// if job timeout and requeueIfTimeout is enable, just continue without
				// requeuing the job.
				_, hasTimeout := err.(*ErrJobTimeout)
				if !q.requeueIfTimeout && hasTimeout {
					q.dropHandler(wk, job)
					continue
				}

				job.nretry++
				if job.nretry <= job.MaxRetry {
					// avoid blocking a worker for delay retrieving
					go func() {
						defer func() { recover() }() // panic if job channel was closed
						time.Sleep(q.retryDelay)
						q.jobq <- job
					}()
				} else {
					q.dropHandler(wk, job)
				}
			}

			if !consumed {
				q.dropHandler(nil, job) // drop the job if no worker can consume it
			}

		case _, ok := <-wkch.suspend:
			if ok {
				<-wkch.resume
			}
		case <-wkch.resume:
			// ignore resume signal if not suspended
			continue
		case <-wkch.terminate:
			return
		}
	}
}
