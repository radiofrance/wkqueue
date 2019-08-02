package jobqueue

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNew(t *testing.T) {
	/*
		worker := WorkerFunc(func(Job) error { return nil })
		successH := func(Worker, Job) {}
		dropH := func(Worker, Job) {}
		errH := func(Worker, error, Job) {}
		panicH := func(Worker, interface{}, Job) {}

		defaultsOpts := []Options{AddWorker(worker), SetSuccessHandler(successH), SetErrHandler(errH), SetDropHandler(dropH), SetPanicHandler(panicH)}
			qdefault := queue{
				rootWorkers:      []Worker{worker},
				jobq:             make(chan Job, 16),
				workerq:          make(chan workerSocket, runtime.NumCPU()),
				jobTimeout:       time.Second,
				retryDelay:       time.Second,
				requeueIfTimeout: false,
				succeedHandler:   successH,
				dropHandler:      dropH,
				errHandler:       errH,
				panicHandler:     panicH,
				sync:             sync.RWMutex{},
				suspended:        false,
				closed:           false,
			}
	*/
	tests := []struct {
		name   string
		opts   []Options
		expect *queue
		err    error
	}{
		{name: "Err_NoWorker", err: fmt.Errorf(errNoWorker)},
		{name: "Err_NilWorker", opts: []Options{AddWorker(nil)}, err: fmt.Errorf(errNilWorker)},
		{name: "Err_NilHandler_Success", opts: []Options{SetSuccessHandler(nil)}, err: fmt.Errorf(errNilHandler, "SuccessHandler")},

		// TODO: Cannot compare directly because new channel will be created
		/*		{
					name: "Set_JobCapacity",
					opts: append(defaultsOpts, []Options{AddWorker(worker), SetJobCapacity(1024)}...),
					expect: &queue{
						rootWorkers:      []Worker{worker},
						jobq:             make(chan Job, 1024),
						workerq:          make(chan workerSocket, runtime.NumCPU()),
						jobTimeout:       time.Second,
						retryDelay:       time.Second,
						requeueIfTimeout: false,
						succeedHandler:   successH,
						dropHandler:      dropH,
						errHandler:       errH,
						panicHandler:     panicH,
						sync:             sync.RWMutex{},
						suspended:        false,
						closed:           false,
					},
				},
		*/}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(tt.opts...)
			if tt.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, q)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func TestNew_Default(t *testing.T) {
	jq, err := New(AddWorker(WorkerFunc(func(Job) error { return nil })))
	assert.NoError(t, err)

	assert.Equal(t, uint(runtime.NumCPU()), jq.WorkersLimit())
	assert.Equal(t, uint(16), jq.JobCapacity())

	jqq := jq.(*queue)
	assert.Equal(t, time.Second, jqq.jobTimeout)
	assert.Equal(t, time.Second, jqq.retryDelay)
	assert.Equal(t, func(Worker, Job) { return }, jqq.succeedHandler)
	assert.Equal(t, func(Worker, Job) { return }, jqq.dropHandler)
	assert.Equal(t, func(Worker, error, Job) { return }, jqq.errHandler)
	assert.Equal(t, func(Worker, interface{}, Job) { return }, jqq.panicHandler)
}

type JobQueueTestSuite struct {
	suite.Suite
	jq   Queue
	jobs chan Job
}

func (s *JobQueueTestSuite) SetupTest() {
	s.jobs = make(chan Job, 16)
	s.jq, _ = New(
		AddWorker(WorkerFunc(func(j Job) error { s.jobs <- j; return nil })),
		SetWorkerCapacity(1),
	)
}
