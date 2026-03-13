package storage

import (
	"context"
	"time"
)

// IncidentState represents the lifecycle state of an incident.
type IncidentState int

const (
	IncidentCreated IncidentState = iota
	IncidentResolved
)

// IncidentType categorizes what caused the incident.
type IncidentType string

const (
	IncidentTypeJobNoRetries       IncidentType = "JOB_NO_RETRIES"
	IncidentTypeConditionError     IncidentType = "CONDITION_ERROR"
	IncidentTypeIOMappingError     IncidentType = "IO_MAPPING_ERROR"
	IncidentTypeMessageCorrelation IncidentType = "MESSAGE_CORRELATION"
	IncidentTypeIntentProcessing   IncidentType = "INTENT_PROCESSING"
)

// Incident is a mutable entity representing a problem that requires attention.
type Incident struct {
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the affected process instance.
	ProcessInstanceKey uint64

	// ElementInstanceKey links to the affected element instance.
	ElementInstanceKey uint64

	// JobKey links to the affected job (0 if not job-related).
	JobKey uint64

	// Type categorizes the incident.
	Type IncidentType

	// State is the current lifecycle state.
	State IncidentState

	// ErrorMessage describes what went wrong.
	ErrorMessage string

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time

	// ResolvedAt is when the incident was resolved (zero if not resolved).
	ResolvedAt time.Time
}

// IncidentRepository manages incident storage.
type IncidentRepository interface {
	// Create stores a new incident.
	Create(ctx context.Context, incident *Incident) error

	// GetByKey retrieves an incident by its key.
	GetByKey(ctx context.Context, key uint64) (*Incident, error)

	// Resolve marks an incident as resolved.
	Resolve(ctx context.Context, key uint64) error

	// FindByProcessInstance returns all incidents for a process instance.
	FindByProcessInstance(ctx context.Context, piKey uint64) ([]*Incident, error)

	// FindUnresolved returns all unresolved incidents.
	FindUnresolved(ctx context.Context, limit int) ([]*Incident, error)
}
