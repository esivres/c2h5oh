package storage

import (
	"context"
	"time"
)

// ProcessInstanceState represents the lifecycle state of a process instance.
type ProcessInstanceState int

const (
	ProcessInstanceCreated ProcessInstanceState = iota
	ProcessInstanceActive
	ProcessInstanceCompleted
	ProcessInstanceTerminated
)

// ProcessInstance is a mutable entity representing a running process.
type ProcessInstance struct {
	CreatedAt            time.Time
	BpmnProcessId        string
	Key                  uint64
	ProcessDefinitionKey uint64
	ParentKey            uint64
	ParentElementKey     uint64
	State                ProcessInstanceState
}

// ElementInstanceState represents the lifecycle of an element instance (token).
type ElementInstanceState int

const (
	ElementInstanceActivating ElementInstanceState = iota
	ElementInstanceActivated
	ElementInstanceCompleting
	ElementInstanceCompleted
	ElementInstanceTerminating
	ElementInstanceTerminated
)

// ElementInstance is a mutable entity representing a token on a BPMN element.
type ElementInstance struct {
	CreatedAt            time.Time
	ElementId            string
	ElementType          string
	Key                  uint64
	ProcessInstanceKey   uint64
	ProcessDefinitionKey uint64
	FlowScopeKey         uint64
	State                ElementInstanceState
}

// ProcessInstanceRepository manages process instance and element instance storage.
type ProcessInstanceRepository interface {
	// CreateInstance stores a new process instance.
	CreateInstance(ctx context.Context, pi *ProcessInstance) error

	// GetInstance retrieves a process instance by key.
	GetInstance(ctx context.Context, key uint64) (*ProcessInstance, error)

	// UpdateInstanceState updates the state of a process instance.
	UpdateInstanceState(ctx context.Context, key uint64, state ProcessInstanceState) error

	// CreateElementInstance stores a new element instance.
	CreateElementInstance(ctx context.Context, ei *ElementInstance) error

	// GetElementInstance retrieves an element instance by key.
	GetElementInstance(ctx context.Context, key uint64) (*ElementInstance, error)

	// UpdateElementInstanceState updates the state of an element instance.
	UpdateElementInstanceState(ctx context.Context, key uint64, state ElementInstanceState) error

	// FindElementInstancesByProcessInstance returns all element instances for a process instance.
	FindElementInstancesByProcessInstance(ctx context.Context, piKey uint64) ([]*ElementInstance, error)
}
