package jobqueue

import (
	"fmt"
	"time"
)

func ExampleNew() {
	q, _ := New(
		AddWorker(WorkerFunc(func(job Job) error { fmt.Printf("Do: %v\n", job.Headers); return nil })),

		SetJobCapacity(2),
		SetWorkerCapacity(1),
		RequeueIfTimeout(),

		SetSuccessHandler(func(worker Worker, job Job) { fmt.Printf("Success: %v\n", job.Headers) }),
	)

	q.Scale(1)

	q.Sync() <- Job{Headers: map[string]interface{}{"id": 1}}
	q.Sync() <- Job{Headers: map[string]interface{}{"id": 2}}
	q.Sync() <- Job{Headers: map[string]interface{}{"id": 3}}
	q.Sync() <- Job{Headers: map[string]interface{}{"id": 4}}

	q.SuspendWorkers()
	q.Async(500 * time.Microsecond) <- Job{Headers: map[string]interface{}{"id": 5}}
	q.Async(500 * time.Microsecond) <- Job{Headers: map[string]interface{}{"id": 6}}
	q.Async(500 * time.Microsecond) <- Job{Headers: map[string]interface{}{"id": 7}}
	q.Async(500 * time.Microsecond) <- Job{Headers: map[string]interface{}{"id": 8}}
	q.ResumeWorkers()

	q.WaitAndClose()
}
