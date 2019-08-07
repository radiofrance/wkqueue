package jobqueue

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

var testmap = map[string]func(t *testing.T, q *queue, expect interface{}){
	"rootWorkers": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, q.rootWorkers, "invalid value for 'rootWorkers'")
	},
	"jobq": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, cap(q.jobq), "invalid value for 'jobq'")
	},
	"workerq": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, cap(q.workerq), "invalid value for 'workerq'")
	},
	"jobTimeout": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, q.jobTimeout, "invalid value for 'jobTimeout'")
	},
	"retryDelay": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, q.retryDelay, "invalid value for 'retryDelay'")
	},
	"requeueIfTimeout": func(t *testing.T, q *queue, expect interface{}) {
		assert.Equal(t, expect, q.requeueIfTimeout, "invalid value for 'requeueIfTimeout'")
	},
	"successHandler": func(t *testing.T, q *queue, expect interface{}) {
		assert.IsType(t, expect, q.succeedHandler, "invalid value for 'succeedHandler'")
	},
	"dropHandler": func(t *testing.T, q *queue, expect interface{}) {
		assert.IsType(t, expect, q.dropHandler, "invalid value for 'dropHandler'")
	},
	"errHandler": func(t *testing.T, q *queue, expect interface{}) {
		assert.IsType(t, expect, q.errHandler, "invalid value for 'errHandler'")
	},
	"panicHandler": func(t *testing.T, q *queue, expect interface{}) {
		assert.IsType(t, expect, q.panicHandler, "invalid value for 'panicHandler'")
	},
}

func TestNew(t *testing.T) {
	worker := WorkerFunc(func(*Job) error { return nil })
	successH := SuccessHandler(func(Worker, *Job) {})
	dropH := DropHandler(func(Worker, *Job) {})
	errH := ErrHandler(func(Worker, error, *Job) {})
	panicH := PanicHandler(func(Worker, interface{}, *Job) {})

	defaultsOpts := []Options{AddWorker(worker)}

	tests := []struct {
		name   string
		opts   []Options
		expect map[string]interface{}
		err    error
	}{
		{name: "Err_NoWorker", err: fmt.Errorf(errNoWorker)},
		{name: "Err_NilWorker", opts: []Options{AddWorker(nil)}, err: fmt.Errorf(errNilWorker)},
		{name: "Err_NilHandler_Success", opts: []Options{SetSuccessHandler(nil)}, err: fmt.Errorf(errNilHandler, "SuccessHandler")},
		{name: "Err_NilHandler_Drop", opts: []Options{SetDropHandler(nil)}, err: fmt.Errorf(errNilHandler, "DropHandler")},
		{name: "Err_NilHandler_Error", opts: []Options{SetErrHandler(nil)}, err: fmt.Errorf(errNilHandler, "ErrHandler")},
		{name: "Err_NilHandler_Panic", opts: []Options{SetPanicHandler(nil)}, err: fmt.Errorf(errNilHandler, "PanicHandler")},

		{name: "Set_JobCapacity", opts: append(defaultsOpts, SetJobCapacity(1024)), expect: map[string]interface{}{"jobq": 1024}},
		{name: "Set_WorkerCapacity", opts: append(defaultsOpts, SetWorkerCapacity(32)), expect: map[string]interface{}{"workerq": 32}},
		{name: "Set_JobTimeout", opts: append(defaultsOpts, SetJobTimeout(time.Minute)), expect: map[string]interface{}{"jobTimeout": time.Minute}},
		{name: "Set_RetryDelay", opts: append(defaultsOpts, SetRetryDelay(time.Minute)), expect: map[string]interface{}{"retryDelay": time.Minute}},
		{name: "Set_RequeueIfTimeout", opts: append(defaultsOpts, RequeueIfTimeout()), expect: map[string]interface{}{"requeueIfTimeout": true}},
		{name: "Set_SuccessHandler", opts: append(defaultsOpts, SetSuccessHandler(successH)), expect: map[string]interface{}{"successHandler": successH}},
		{name: "Set_DropHandler", opts: append(defaultsOpts, SetDropHandler(dropH)), expect: map[string]interface{}{"dropHandler": dropH}},
		{name: "Set_ErrHandler", opts: append(defaultsOpts, SetErrHandler(errH)), expect: map[string]interface{}{"errHandler": errH}},
		{name: "Set_PanicHandler", opts: append(defaultsOpts, SetPanicHandler(panicH)), expect: map[string]interface{}{"panicHandler": panicH}},

		{
			name:   "Set_JobCapacity_PanicHandler",
			opts:   append(defaultsOpts, SetJobCapacity(12), SetPanicHandler(panicH)),
			expect: map[string]interface{}{"jobq": 12, "panicHandler": panicH},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(tt.opts...)
			if tt.err == nil {
				assert.NoError(t, err)
				for k, expect := range tt.expect {
					testmap[k](t, q.(*queue), expect)
				}
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func TestNew_Default(t *testing.T) {
	jq, err := New(AddWorker(WorkerFunc(func(*Job) error { return nil })))
	assert.NoError(t, err)

	assert.Equal(t, runtime.NumCPU(), jq.WorkersLimit())
	assert.Equal(t, 16, jq.JobCapacity())

	jqq := jq.(*queue)
	assert.Equal(t, time.Second, jqq.jobTimeout)
	assert.Equal(t, time.Second, jqq.retryDelay)
	assert.NotNil(t, jqq.succeedHandler)
	assert.NotNil(t, jqq.dropHandler)
	assert.NotNil(t, jqq.errHandler)
	assert.NotNil(t, jqq.panicHandler)
}

func TestJobQueueTestSuite(t *testing.T) { suite.Run(t, new(JobQueueTestSuite)) }
func TestWorkerTestSuite(t *testing.T)   { suite.Run(t, new(WorkerTestSuite)) }
func TestJobTestSuite(t *testing.T)      { suite.Run(t, new(JobTestSuite)) }

type JobQueueTestSuite struct {
	suite.Suite
	q    Queue
	jobs chan Job
}

func (s *JobQueueTestSuite) SetupTest() {
	s.q, _ = New(
		AddWorker(WorkerFunc(func(*Job) error { return nil })),
		SetWorkerCapacity(4),
	)
}
func (s *JobQueueTestSuite) TearDown() {
	s.q.Close()
	close(s.jobs)
}

func (s *JobQueueTestSuite) TestSync() {
	// fill job queue
	for i := 0; i < s.q.JobCapacity(); i++ {
		s.q.Sync() <- &Job{MaxRetry: 1}
	}
	s.Equal(s.q.JobCapacity(), s.q.JobLoad())

	// sync test when job queue is full
	select {
	case s.q.Sync() <- &Job{MaxRetry: 1}:
		s.FailNow("Sync call must be blocked (job queue is full)")
	case <-time.After(time.Millisecond):
	}

	// remove some jobs
	s.q.Scale(4)
	time.Sleep(time.Millisecond)
	s.NotEqual(s.q.JobCapacity(), s.q.JobLoad())

	// sync test when job queue is not full
	select {
	case s.q.Sync() <- &Job{MaxRetry: 1}:
	case <-time.After(time.Millisecond):
		s.FailNow("Sync call must not be blocked (job queue is not full)")
	}

	// call sync after close jobq must be ignored (to avoid panic)
	s.q.Close()
	s.NotPanics(func() { s.q.Sync() <- &Job{MaxRetry: 1} })
}
func (s *JobQueueTestSuite) TestMultiClose() {
	s.q.Close()
	s.q.Close()
	s.q.WaitAndClose()
}
func (s *JobQueueTestSuite) TestClose() {
	s.q.Close()
	s.q, _ = New(
		AddWorker(WorkerFunc(func(*Job) error { time.Sleep(time.Millisecond); return nil })),
	)
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}

	s.q.Scale(4)
	start := time.Now()
	s.q.Close()
	end := time.Now()

	// close will be instantaneous because the job queue is flushed
	s.Less(int64(end.Sub(start)), int64(time.Millisecond))
}
func (s *JobQueueTestSuite) TestWaitAndClose() {
	s.q.Close()
	s.q, _ = New(
		AddWorker(WorkerFunc(func(*Job) error { time.Sleep(time.Millisecond); return nil })),
	)
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}
	s.q.Sync() <- &Job{MaxRetry: 1}

	s.q.Scale(4)
	start := time.Now()
	s.q.WaitAndClose()
	end := time.Now()

	// close will not be instantaneous because it waits for all the work to be consumed before being closed
	s.GreaterOrEqual(int64(end.Sub(start)), int64(time.Millisecond))
}

type WorkerTestSuite struct{ suite.Suite }

func (s *WorkerTestSuite) TestWorkerScaling() {
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { return nil })),
		SetWorkerCapacity(4),
	)
	defer q.Close()

	tests := []struct {
		scaleTo    uint
		err        error
		delta      int
		numWorkers int
	}{
		{scaleTo: 2, delta: 2, numWorkers: 2},
		{scaleTo: 3, delta: 1, numWorkers: 3},
		{scaleTo: 5, err: fmt.Errorf(errMaxWorkerReached), delta: 1, numWorkers: 4},
		{scaleTo: 1, delta: -3, numWorkers: 1},
		{scaleTo: 1, delta: 0, numWorkers: 1},
		{scaleTo: 0, delta: -1, numWorkers: 0},
	}

	for _, tt := range tests {
		n, err := q.Scale(tt.scaleTo)
		if tt.err != nil {
			s.EqualError(err, tt.err.Error())
		} else {
			s.NoError(err)
		}
		s.Equal(tt.delta, n)
		s.Equal(tt.numWorkers, q.NumWorkers())
	}
}
func (s *WorkerTestSuite) TestFailedWorker() {
	q, _ := New(AddWorker(&errWorker{}))
	defer q.Close()

	n, err := q.Scale(1)
	s.EqualError(err, "nope")
	s.Equal(0, n)
}
func (s *WorkerTestSuite) TestMultiWorker() {
	wk1, wk2 := make(chan *Job), make(chan *Job)
	q, _ := New(
		AddWorker(WorkerFunc(func(j *Job) error { wk1 <- j; return nil })),
		AddWorker(WorkerFunc(func(j *Job) error { wk2 <- j; return nil })),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 1}
	s.Equal(&Job{MaxRetry: 1}, <-wk1)
	s.Equal(&Job{MaxRetry: 1}, <-wk2)
}
func (s *WorkerTestSuite) TestSuspendWorker() {
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { return nil })),
		SetJobCapacity(1),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{}
	time.Sleep(time.Millisecond) // wait worker to consume the job
	s.Equal(0, q.JobLoad())

	q.SuspendWorkers()

	q.Sync() <- &Job{}
	time.Sleep(time.Millisecond) // wait worker to consume the job
	s.Equal(1, q.JobLoad())      // expect 1 because workers are suspended

	q.ResumeWorkers()

	time.Sleep(time.Millisecond) // wait worker to consume the job
	s.Equal(0, q.JobLoad())
}

type JobTestSuite struct{ suite.Suite }

func (s *JobTestSuite) getDroppedJob(ch chan *Job) *Job {
	select {
	case j := <-ch:
		return j
	case <-time.After(time.Second):
		s.FailNow("message not dropped")
	}
	return &Job{}
}
func (s *JobTestSuite) getPanickedJob(ch chan interface{}) interface{} {
	select {
	case x := <-ch:
		return x
	case <-time.After(time.Second):
		s.FailNow("worker not panicked")
	}
	return nil
}
func (s *JobTestSuite) TestJobDrop() {
	droppedJob := make(chan *Job)
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { return xerrors.New("drop it !") })),
		SetRetryDelay(0),
		SetDropHandler(func(_ Worker, job *Job) { droppedJob <- job }),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 1}
	s.Equal(2, s.getDroppedJob(droppedJob).nretry)

	q.Sync() <- &Job{MaxRetry: 5}
	s.Equal(6, s.getDroppedJob(droppedJob).nretry)
}
func (s *JobTestSuite) TestDropNoValidWorker() {
	droppedJob := make(chan *Job)
	q, _ := New(
		AddWorker(&dummyWorker{}),
		SetRetryDelay(0),
		SetDropHandler(func(_ Worker, job *Job) { droppedJob <- job }),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 5}
	s.Equal(0, s.getDroppedJob(droppedJob).nretry)
}
func (s *JobTestSuite) TestJobTimeout() {
	droppedJob := make(chan *Job)
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { time.Sleep(time.Microsecond); return nil })),
		SetJobTimeout(time.Nanosecond),
		SetRetryDelay(0),
		SetDropHandler(func(_ Worker, job *Job) { droppedJob <- job }),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 5}
	s.Equal(0, s.getDroppedJob(droppedJob).nretry)
}
func (s *JobTestSuite) TestJobRequeueTimeout() {
	droppedJob := make(chan *Job)
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { time.Sleep(time.Microsecond); return nil })),
		RequeueIfTimeout(),
		SetJobTimeout(time.Nanosecond),
		SetRetryDelay(0),
		SetDropHandler(func(_ Worker, job *Job) { droppedJob <- job }),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 5}
	s.Equal(6, s.getDroppedJob(droppedJob).nretry)
}
func (s *JobTestSuite) TestJobPanic() {
	droppedJob := make(chan *Job)
	panickedJob := make(chan interface{})
	q, _ := New(
		AddWorker(WorkerFunc(func(*Job) error { panic("wooops") })),
		RequeueIfTimeout(),
		SetJobTimeout(time.Microsecond),
		SetRetryDelay(0),
		SetDropHandler(func(_ Worker, job *Job) { droppedJob <- job }),
		SetPanicHandler(func(_ Worker, r interface{}, job *Job) { panickedJob <- r }),
	)
	defer q.Close()
	q.Scale(1)

	q.Sync() <- &Job{MaxRetry: 5}
	s.Equal("wooops", s.getPanickedJob(panickedJob))
	s.Equal("wooops", s.getPanickedJob(panickedJob))
	s.Equal("wooops", s.getPanickedJob(panickedJob))
	s.Equal("wooops", s.getPanickedJob(panickedJob))
	s.Equal("wooops", s.getPanickedJob(panickedJob))
	s.Equal(6, s.getDroppedJob(droppedJob).nretry)
}

type dummyWorker struct{}

func (d dummyWorker) Initialize() error       { return nil }
func (d dummyWorker) Terminate()              {}
func (d dummyWorker) Do(job *Job) error        { return nil }
func (d dummyWorker) CanConsume(job *Job) bool { return job.Headers != nil }
func (d dummyWorker) Copy() Worker            { return &dummyWorker{} }

type errWorker struct{}

func (e errWorker) Initialize() error       { return xerrors.New("nope") }
func (e errWorker) Terminate()              {}
func (e errWorker) Do(job *Job) error        { return nil }
func (e errWorker) CanConsume(job *Job) bool { return true }
func (e errWorker) Copy() Worker            { return &errWorker{} }
