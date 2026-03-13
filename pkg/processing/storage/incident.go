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
	CreatedAt          time.Time
	ResolvedAt         time.Time
	Type               IncidentType
	ErrorMessage       string
	Key                uint64
	ProcessInstanceKey uint64
	ElementInstanceKey uint64
	JobKey             uint64
	State              IncidentState
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
