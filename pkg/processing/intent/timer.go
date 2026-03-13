package intent

import "time"

// CreateTimerIntent requests creation of a timer.
type CreateTimerIntent struct {
	DueDate time.Time
	Header
	ElementInstanceKey   uint64
	ProcessDefinitionKey uint64
	Repetitions          int
	CycleDuration        time.Duration
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
