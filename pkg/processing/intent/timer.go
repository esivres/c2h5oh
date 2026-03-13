package intent

import "time"

// CreateTimerIntent requests creation of a timer.
type CreateTimerIntent struct {
	Header

	// ElementInstanceKey is the element instance that owns this timer.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// DueDate is when the timer should trigger.
	DueDate time.Time

	// Repetitions is the number of times to repeat (-1 for infinite, 0 for one-shot).
	Repetitions int

	// CycleDuration is the interval between repetitions (used for timer cycle).
	CycleDuration time.Duration
}

func (i *CreateTimerIntent) IntentType() Type { return CreateTimer }

// TriggerTimerIntent signals that a timer has reached its due date.
type TriggerTimerIntent struct {
	Header

	// TimerKey is the key of the triggered timer.
	TimerKey uint64
}

func (i *TriggerTimerIntent) IntentType() Type { return TriggerTimer }

// CancelTimerIntent requests cancellation of a timer.
type CancelTimerIntent struct {
	Header

	// TimerKey is the key of the timer to cancel.
	TimerKey uint64
}

func (i *CancelTimerIntent) IntentType() Type { return CancelTimer }
