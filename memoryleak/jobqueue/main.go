package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/radiofrance/wkqueue/jobqueue"
)

type BigObject [2 * 1024 * 1024]byte

func producer(q jobqueue.Queue) {
	for {
		q.Sync() <- &jobqueue.Job{Payload: BigObject{}}
		time.Sleep(time.Millisecond)
	}
}

func watchdog(q jobqueue.Queue) {
	for {
		time.Sleep(250 * time.Millisecond)
		fmt.Printf("(watchdog) > %d in queue\n", q.JobLoad())
	}
}

func chaos(q jobqueue.Queue) {
	for {
		time.Sleep(100 * time.Millisecond)
		n := rand.Intn(20)
		switch n {
		case 6, 14:
			fmt.Printf("(chaos) > suspend workers\n")
			q.SuspendWorkers()
		case 3, 17:
			fmt.Printf("(chaos) > resume workers\n")
			q.ResumeWorkers()
		}
	}
}

func main() {
	q, _ := jobqueue.New(
		jobqueue.AddWorker(jobqueue.WorkerFunc(func(j *jobqueue.Job) error { return nil })),
		jobqueue.SetWorkerCapacity(1),
		jobqueue.SetJobCapacity(512),
	)

	q.Scale(1)

	go producer(q)
	go watchdog(q)
	go chaos(q)

	panic(http.ListenAndServe(":6060", nil))
}
