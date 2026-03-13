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
	// Key is the unique identifier.
	Key uint64

	// ProcessDefinitionKey links to the deployed process definition.
	ProcessDefinitionKey uint64

	// BpmnProcessId is the BPMN process id (denormalized for queries).
	BpmnProcessId string

	// ParentKey is the parent process instance key (0 if top-level).
	ParentKey uint64

	// ParentElementKey is the element instance key in the parent that spawned this instance.
	ParentElementKey uint64

	// State is the current lifecycle state.
	State ProcessInstanceState

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
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
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the owning process instance.
	ProcessInstanceKey uint64

	// ProcessDefinitionKey links to the process definition.
	ProcessDefinitionKey uint64

	// ElementId is the BPMN element id (e.g. "Activity_01ran10").
	ElementId string

	// ElementType is the BPMN element type (e.g. "serviceTask", "exclusiveGateway").
	ElementType string

	// FlowScopeKey is the key of the parent scope element instance (process or subprocess).
	FlowScopeKey uint64

	// State is the current lifecycle state.
	State ElementInstanceState

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time
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
