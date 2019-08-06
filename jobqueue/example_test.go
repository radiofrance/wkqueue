package jobqueue_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/radiofrance/wkqueue/jobqueue"
)

// DummyWorker implement a simple worker.
type DummyWorker struct{ string }

func (d *DummyWorker) Initialize() error {
	d.string = fmt.Sprintf("%d", rand.Intn(100))
	return nil
}
func (d DummyWorker) Terminate() {}
func (d DummyWorker) Do(job jobqueue.Job) error {
	fmt.Printf("(DummyWorker#%s) => %#v\n", d.string, job)
	return nil
}
func (d DummyWorker) CanConsume(job jobqueue.Job) bool { return job.Headers.Has("dummy.enabled") }
func (d DummyWorker) Copy() jobqueue.Worker            { return &DummyWorker{} }

func ExampleQueue_simpleQueue() {
	q, _ := jobqueue.New(
		jobqueue.AddWorker(jobqueue.WorkerFunc(func(j jobqueue.Job) error {
			fmt.Printf("(WorkerFunc) => %#v\n", j)
			return nil
		})),
		jobqueue.SetJobCapacity(1),
		jobqueue.SetWorkerCapacity(2),
	)

	fmt.Println("(main) => Upscale workers")
	_, _ = q.Scale(2)

	fmt.Println("(main) => Fill job queue")
	q.Sync() <- jobqueue.Job{Payload: 0}
	q.Sync() <- jobqueue.Job{Payload: 1}
	q.Sync() <- jobqueue.Job{Payload: 2}
	q.Sync() <- jobqueue.Job{Payload: 3}
	q.Sync() <- jobqueue.Job{Payload: 4}
	q.Sync() <- jobqueue.Job{Payload: 5}

	q.WaitAndClose()
}

func ExampleQueue_multiWorkerQueue() {
	q, _ := jobqueue.New(
		jobqueue.AddWorker(jobqueue.WorkerFunc(func(j jobqueue.Job) error {
			fmt.Printf("(WorkerFunc) => %#v\n", j)
			return nil
		})),
		jobqueue.AddWorker(&DummyWorker{}),
		jobqueue.SetJobCapacity(1),
		jobqueue.SetWorkerCapacity(2),
	)

	fmt.Println("(main) => Upscale workers")
	_, _ = q.Scale(2)

	fmt.Println("(main) => Fill job queue")
	q.Sync() <- jobqueue.Job{Payload: 0}
	q.Sync() <- jobqueue.Job{Payload: 1, Headers: map[string]interface{}{"dummy.enabled": true}}
	q.Sync() <- jobqueue.Job{Payload: 2}
	q.Sync() <- jobqueue.Job{Payload: 3}
	q.Sync() <- jobqueue.Job{Payload: 4, Headers: map[string]interface{}{"dummy.enabled": true}}
	q.Sync() <- jobqueue.Job{Payload: 5}

	q.WaitAndClose()
}

func ExampleQueue_suspendedQueue() {
	q, _ := jobqueue.New(
		jobqueue.AddWorker(jobqueue.WorkerFunc(func(j jobqueue.Job) error {
			fmt.Printf("(WorkerFunc) => %#v\n", j)
			return nil
		})),
		jobqueue.SetWorkerCapacity(8),
	)

	fmt.Println("(main) => Upscale workers")
	_, _ = q.Scale(8)

	q.Sync() <- jobqueue.Job{Payload: 0}
	q.Sync() <- jobqueue.Job{Payload: 1}
	q.Sync() <- jobqueue.Job{Payload: 2}

	fmt.Println("(main) => Wait 100ms")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("(main) => Suspend workers")
	q.SuspendWorkers()

	q.Sync() <- jobqueue.Job{Payload: 3}
	q.Sync() <- jobqueue.Job{Payload: 4}
	q.Sync() <- jobqueue.Job{Payload: 5}
	q.Sync() <- jobqueue.Job{Payload: 6}
	q.Sync() <- jobqueue.Job{Payload: 7}
	q.Sync() <- jobqueue.Job{Payload: 8}
	q.Sync() <- jobqueue.Job{Payload: 9}

	fmt.Printf("(main) => job queue has %d elements\n", q.JobLoad())
	fmt.Println("(main) => Downscale workers")
	_, _ = q.Scale(3)

	fmt.Println("(main) => Resume workers")
	q.ResumeWorkers()

	q.WaitAndClose()
}
