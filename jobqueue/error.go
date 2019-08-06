package jobqueue

import (
	"golang.org/x/xerrors"
)

// ErrNoWorker occurs when no worker is provided.
type ErrNoWorker struct{ error }

// ErrNilWorker occurs when a nil worker is provided.
type ErrNilWorker struct{ error }

// ErrNilHandler occurs when a nil handler is provided.
type ErrNilHandler struct{ error }

// ErrMaxWorkerReached occurs when the queue reach the maximum number of workers.
type ErrMaxWorkerReached struct{ error }

// ErrMinWorkerReached never occurs (theoretically)
type ErrMinWorkerReached struct{ error }

// ErrJobTimeout occurs when a job time out.
type ErrJobTimeout struct{ error }

// ErrJobPanic occurs when a job panics.
type ErrJobPanic struct{ error }

const (
	errNoWorker         = "at least one worker must be provided"
	errNilWorker        = "worker cannot be nil"
	errNilHandler       = "%s cannot be nil"
	errMaxWorkerReached = "maximum number of workers reached"
	errMinWorkerReached = "minimum number of workers reached"
	errJobTimeout       = "job has timed out"
	errJobPanic         = "job has panicked"
)

func newErrNoWorker() error           { return &ErrNoWorker{xerrors.New(errNoWorker)} }
func newErrNilWorker() error          { return &ErrNilWorker{xerrors.New(errNilWorker)} }
func newErrNilHandler(h string) error { return &ErrNilHandler{xerrors.Errorf(errNilHandler, h)} }
func newErrMaxWorkerReached() error   { return &ErrMaxWorkerReached{xerrors.New(errMaxWorkerReached)} }
func newErrMinWorkerReached() error   { return &ErrMinWorkerReached{xerrors.New(errMinWorkerReached)} }
func newErrJobTimeout() error         { return &ErrJobTimeout{xerrors.New(errJobTimeout)} }
func newErrJobPanic() error           { return &ErrJobPanic{xerrors.New(errJobPanic)} }
