package storage

import (
	"context"
	"time"
)

// TimerState represents the lifecycle state of a timer.
type TimerState int

const (
	TimerCreated TimerState = iota
	TimerTriggered
	TimerCanceled
)

// Timer is a mutable entity representing a scheduled timer event.
type Timer struct {
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the owning process instance.
	ProcessInstanceKey uint64

	// ElementInstanceKey links to the element instance that created this timer.
	ElementInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// State is the current lifecycle state.
	State TimerState

	// DueDate is when the timer should trigger.
	DueDate time.Time

	// Repetitions is the remaining repetition count (-1 for infinite, 0 for one-shot).
	Repetitions int

	// CycleDuration is the interval between repetitions.
	CycleDuration time.Duration

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
}

// TimerRepository manages timer storage.
type TimerRepository interface {
	// Create stores a new timer.
	Create(ctx context.Context, timer *Timer) error

	// GetByKey retrieves a timer by its key.
	GetByKey(ctx context.Context, key uint64) (*Timer, error)

	// Trigger marks a timer as triggered.
	Trigger(ctx context.Context, key uint64) error

	// Cancel marks a timer as canceled.
	Cancel(ctx context.Context, key uint64) error

	// FindDue returns all timers with DueDate <= now in Created state.
	FindDue(ctx context.Context, now time.Time, limit int) ([]*Timer, error)

	// FindByProcessInstance returns all timers for a process instance.
	FindByProcessInstance(ctx context.Context, piKey uint64) ([]*Timer, error)
}
