package jobqueue

import (
	"golang.org/x/xerrors"
)

type ErrNoWorker struct{ error }
type ErrNilWorker struct{ error }
type ErrNilHandler struct{ error }
type ErrMaxWorkerReached struct{ error }
type ErrMinWorkerReached struct{ error }
type ErrJobTimeout struct{ error }
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
