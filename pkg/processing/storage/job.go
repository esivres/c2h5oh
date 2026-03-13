package storage

import (
	"context"
	"time"
)

// JobState represents the lifecycle state of a job.
type JobState int

const (
	JobCreated JobState = iota
	JobActivated
	JobCompleted
	JobFailed
	JobCanceled
	JobErrorThrown
)

// Job is a mutable entity representing work to be done by an external worker.
type Job struct {
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the owning process instance.
	ProcessInstanceKey uint64

	// ElementInstanceKey links to the element instance that created this job.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// Type is the job type used for worker activation (e.g. "payment-service").
	Type string

	// State is the current lifecycle state.
	State JobState

	// Retries is the remaining retry count.
	Retries int

	// Worker is the name of the worker that activated this job (empty if not activated).
	Worker string

	// Deadline is the time by which the job must be completed once activated.
	Deadline time.Time

	// ErrorMessage is set when the job fails.
	ErrorMessage string

	// ErrorCode is set when the job throws a BPMN error.
	ErrorCode string

	// Variables is the serialized job variables (JSON bytes).
	Variables []byte

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
}

// JobRepository manages job storage.
type JobRepository interface {
	// Create stores a new job.
	Create(ctx context.Context, job *Job) error

	// GetByKey retrieves a job by its key.
	GetByKey(ctx context.Context, key uint64) (*Job, error)

	// Activate marks a job as activated by a worker.
	// Sets state=Activated, worker name, and deadline.
	Activate(ctx context.Context, key uint64, worker string, deadline time.Time) error

	// Complete marks a job as completed with optional result variables.
	Complete(ctx context.Context, key uint64, variables []byte) error

	// Fail marks a job as failed, decrements retries.
	Fail(ctx context.Context, key uint64, retries int, errorMessage string) error

	// ThrowError marks a job with a BPMN error.
	ThrowError(ctx context.Context, key uint64, errorCode string, errorMessage string) error

	// FindActivatable returns jobs available for activation by type.
	// Only returns jobs in Created state.
	FindActivatable(ctx context.Context, jobType string, maxJobs int) ([]*Job, error)

	// UpdateRetries updates the retry count of a job.
	// If the job is in Failed state and new retries > 0, moves it back to Created state.
	UpdateRetries(ctx context.Context, key uint64, retries int) error

	// UpdateDeadline updates the deadline of an activated job.
	UpdateDeadline(ctx context.Context, key uint64, deadline time.Time) error

	// FindByProcessInstance returns all jobs for a process instance.
	FindByProcessInstance(ctx context.Context, piKey uint64) ([]*Job, error)
}
