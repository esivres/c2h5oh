package intent

import "time"

// CreateJobIntent requests creation of a new job for an external worker.
type CreateJobIntent struct {
	Type      string
	Variables []byte
	Header
	ElementInstanceKey   uint64
	ProcessDefinitionKey uint64
	Retries              int
}

func (i *CreateJobIntent) IntentType() Type { return CreateJob }

// ActivateJobIntent requests activation of a job by a worker.
type ActivateJobIntent struct {
	Worker string
	Header
	JobKey  uint64
	Timeout time.Duration
}

func (i *ActivateJobIntent) IntentType() Type { return ActivateJob }

// CompleteJobIntent signals that a worker has completed a job.
type CompleteJobIntent struct {
	Variables []byte
	Header
	JobKey uint64
}

func (i *CompleteJobIntent) IntentType() Type { return CompleteJob }

// FailJobIntent signals that a worker failed to process a job.
type FailJobIntent struct {
	ErrorMessage string
	Header
	JobKey       uint64
	Retries      int
	RetryBackoff time.Duration
}

func (i *FailJobIntent) IntentType() Type { return FailJob }

// ThrowJobErrorIntent signals that a worker is throwing a BPMN error.
type ThrowJobErrorIntent struct {
	ErrorCode    string
	ErrorMessage string
	Header
	JobKey uint64
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
