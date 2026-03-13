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
	CreatedAt            time.Time
	Deadline             time.Time
	Worker               string
	ErrorCode            string
	Type                 string
	ErrorMessage         string
	Variables            []byte
	ProcessDefinitionKey uint64
	Retries              int
	State                JobState
	Key                  uint64
	ElementInstanceKey   uint64
	ProcessInstanceKey   uint64
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
