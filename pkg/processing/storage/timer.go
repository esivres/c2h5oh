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
	DueDate              time.Time
	CreatedAt            time.Time
	Key                  uint64
	ProcessInstanceKey   uint64
	ElementInstanceKey   uint64
	ProcessDefinitionKey uint64
	State                TimerState
	Repetitions          int
	CycleDuration        time.Duration
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
