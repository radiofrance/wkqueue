package jobqueue

// JobHeaders contains optional headers.
type JobHeaders map[string]interface{}

// JobPayload contains the main payload which will be consumed by a worker.
type JobPayload interface{}

// Job is the container for the job definition. It contains all data and
// metadata used by the worker.
type Job struct {
	Headers JobHeaders
	PayLoad JobPayload

	MaxRetry int
	nretry   int
}

// isZero returns true if job is 'empty'
func (j Job) isZero() bool {
	return j.Headers == nil && j.PayLoad == nil && j.MaxRetry == 0 && j.nretry == 0
}

// Worker is the entity who consume the Job defined in the job queue. A worker
// can only do one thing and must be terminated.
type Worker interface {
	// Initialize 'initialize' the worker; if a connection must be created
	// or something else, it must be done during the initialization step.
	// This method is called only once, when the instance is created.
	// (It can be considered as a constructor in POO).
	Initialize() error
	// Terminate is called when the instance worker will be destroyed. For
	// example, all connexions created during the 'initialization' step
	// must be closed here.
	// (It can be considered as a destructor in POO).
	Terminate()

	// Do consume the job stored in JobQueue.
	Do(job Job) error

	// CanConsume returns true if the worker can consume the given job.
	CanConsume(job Job) bool

	// Copy returns a 'copy' of a worker. Because some workflow need to share
	// something between all worker, a copy can be made here.
	Copy() Worker
}

// WorkerFunc wraps a function in a valid worker.
type WorkerFunc func(job Job) error

// Initialize does nothing (stateless worker)
func (WorkerFunc) Initialize() error { return nil }

// Terminate does nothing (stateless)
func (WorkerFunc) Terminate() { return }

// Do runs the given function.
func (f WorkerFunc) Do(job Job) error { return f(job) }

// CanConsume always returns true.
func (WorkerFunc) CanConsume(job Job) bool { return true }

// Copy returns the same instance (stateless)
func (f WorkerFunc) Copy() Worker { return f }
