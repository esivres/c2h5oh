package engine

import (
	"github.com/esivres/c2h5oh/pkg/processing/intent"
)

// Priority levels for intent processing.
const (
	PriorityHigh   = 0
	PriorityNormal = 1
	PriorityLow    = 2
)

// PriorityFunc determines the priority of an intent.
// Lower value = higher priority.
type PriorityFunc func(intent.Intent) int

// DefaultPriority assigns priority based on intent type.
func DefaultPriority(i intent.Intent) int {
	switch i.IntentType() {
	case intent.DeployProcess,
		intent.CreateProcessInstance,
		intent.CancelProcessInstance,
		intent.CreateJob,
		intent.ActivateJob,
		intent.SetVariables:
		return PriorityHigh

	case intent.ActivateElement,
		intent.CompleteElement,
		intent.TerminateElement,
		intent.CompleteJob,
		intent.FailJob,
		intent.ThrowJobError,
		intent.CreateIncident,
		intent.ResolveIncident:
		return PriorityNormal

	default:
		return PriorityLow
	}
}

// Queue is a priority-based intent queue.
// High priority intents are dequeued before normal and low.
type Queue struct {
	high   chan intent.Intent
	normal chan intent.Intent
	low    chan intent.Intent
	prio   PriorityFunc
}

// NewQueue creates a new priority queue with the given buffer size per level.
func NewQueue(bufferSize int, prio PriorityFunc) *Queue {
	if prio == nil {
		prio = DefaultPriority
	}
	return &Queue{
		high:   make(chan intent.Intent, bufferSize),
		normal: make(chan intent.Intent, bufferSize),
		low:    make(chan intent.Intent, bufferSize),
		prio:   prio,
	}
}

// Enqueue adds an intent to the appropriate priority channel.
// Blocks if the channel is full.
func (q *Queue) Enqueue(i intent.Intent) {
	switch q.prio(i) {
	case PriorityHigh:
		q.high <- i
	case PriorityNormal:
		q.normal <- i
	default:
		q.low <- i
	}
}

// TryEnqueue adds an intent without blocking.
// Returns false if the channel is full.
func (q *Queue) TryEnqueue(i intent.Intent) bool {
	switch q.prio(i) {
	case PriorityHigh:
		select {
		case q.high <- i:
			return true
		default:
			return false
		}
	case PriorityNormal:
		select {
		case q.normal <- i:
			return true
		default:
			return false
		}
	default:
		select {
		case q.low <- i:
			return true
		default:
			return false
		}
	}
}

// Dequeue returns the next intent, preferring higher priority.
// Blocks until an intent is available or ctx channel is closed.
func (q *Queue) Dequeue(done <-chan struct{}) (intent.Intent, bool) {
	// Try high priority first (non-blocking)
	select {
	case i := <-q.high:
		return i, true
	default:
	}

	// Try normal priority (non-blocking)
	select {
	case i := <-q.normal:
		return i, true
	default:
	}

	// Try low priority (non-blocking)
	select {
	case i := <-q.low:
		return i, true
	default:
	}

	// Nothing available — block until something arrives or done
	select {
	case i := <-q.high:
		return i, true
	case i := <-q.normal:
		return i, true
	case i := <-q.low:
		return i, true
	case <-done:
		return nil, false
	}
}
