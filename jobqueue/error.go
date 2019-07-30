package jobqueue

import (
	"golang.org/x/xerrors"
)

type ErrNilWorker error
type ErrNilHandler error
type ErrMaxWorkerReached error
type ErrMinWorkerReached error
type ErrJobTimeout error
type ErrJobPanic error

const (
	errNilWorker        = "worker cannot be nil"
	errNilHandler       = "%s cannot be nil"
	errMaxWorkerReached = "maximum number of workers reached"
	errMinWorkerReached = "minimum number of workers reached"
	errJobTimeout       = "job has timed out"
	errJobPanic         = "job has panicked"
)

func newErrNilWorker() error          { return ErrNilWorker(xerrors.New(errNilWorker)) }
func newErrNilHandler(h string) error { return ErrNilHandler(xerrors.Errorf(errNilWorker, h)) }
func newErrMaxWorkerReached() error   { return ErrMaxWorkerReached(xerrors.New(errMaxWorkerReached)) }
func newErrMinWorkerReached() error   { return ErrMinWorkerReached(xerrors.New(errMinWorkerReached)) }
func newErrJobTimeout() error         { return ErrJobTimeout(xerrors.New(errJobTimeout)) }
func newErrJobPanic() error           { return ErrJobPanic(xerrors.New(errJobPanic)) }
