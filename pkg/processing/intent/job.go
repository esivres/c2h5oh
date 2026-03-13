package intent

import "time"

// CreateJobIntent requests creation of a new job for an external worker.
type CreateJobIntent struct {
	Header

	// ElementInstanceKey is the element instance that spawned this job.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// Type is the job type for worker matching (e.g. "payment-service").
	Type string

	// Retries is the initial retry count.
	Retries int

	// Variables is the job input variables (JSON bytes).
	Variables []byte
}

func (i *CreateJobIntent) IntentType() Type { return CreateJob }

// ActivateJobIntent requests activation of a job by a worker.
type ActivateJobIntent struct {
	Header

	// JobKey is the key of the job to activate.
	JobKey uint64

	// Worker is the name of the activating worker.
	Worker string

	// Timeout is the duration after which the job times out.
	Timeout time.Duration
}

func (i *ActivateJobIntent) IntentType() Type { return ActivateJob }

// CompleteJobIntent signals that a worker has completed a job.
type CompleteJobIntent struct {
	Header

	// JobKey is the key of the completed job.
	JobKey uint64

	// Variables is the output variables from the worker (JSON bytes).
	Variables []byte
}

func (i *CompleteJobIntent) IntentType() Type { return CompleteJob }

// FailJobIntent signals that a worker failed to process a job.
type FailJobIntent struct {
	Header

	// JobKey is the key of the failed job.
	JobKey uint64

	// Retries is the remaining retry count.
	Retries int

	// ErrorMessage describes what went wrong.
	ErrorMessage string

	// RetryBackoff is the delay before retrying.
	RetryBackoff time.Duration
}

func (i *FailJobIntent) IntentType() Type { return FailJob }

// ThrowJobErrorIntent signals that a worker is throwing a BPMN error.
type ThrowJobErrorIntent struct {
	Header

	// JobKey is the key of the job.
	JobKey uint64

	// ErrorCode is the BPMN error code.
	ErrorCode string

	// ErrorMessage is the human-readable message.
	ErrorMessage string
}

func (i *ThrowJobErrorIntent) IntentType() Type { return ThrowJobError }

// TimeOutJobIntent signals that an activated job has exceeded its deadline.
type TimeOutJobIntent struct {
	Header

	// JobKey is the key of the timed-out job.
	JobKey uint64
}

func (i *TimeOutJobIntent) IntentType() Type { return TimeOutJob }

// UpdateJobRetriesIntent requests updating the retry count of a job.
type UpdateJobRetriesIntent struct {
	Header

	// JobKey is the key of the job to update.
	JobKey uint64

	// Retries is the new retry count.
	Retries int
}

func (i *UpdateJobRetriesIntent) IntentType() Type { return UpdateJobRetries }

// UpdateJobTimeoutIntent requests updating the deadline of an activated job.
type UpdateJobTimeoutIntent struct {
	Header

	// JobKey is the key of the job to update.
	JobKey uint64

	// Timeout is the new timeout duration (deadline = now + timeout).
	Timeout time.Duration
}

func (i *UpdateJobTimeoutIntent) IntentType() Type { return UpdateJobTimeout }
