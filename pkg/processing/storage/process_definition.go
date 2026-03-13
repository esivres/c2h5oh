package storage

import (
	"context"
	"time"
)

// ProcessDefinition is an immutable entity representing a deployed BPMN process.
// Once created, it is never modified.
type ProcessDefinition struct {
	DeployedAt    time.Time
	BpmnProcessId string
	Name          string
	ContentHash   []byte
	Content       []byte
	Key           uint64
	Version       uint64
}

// ProcessDefinitionRepository manages process definition storage.
type ProcessDefinitionRepository interface {
	// Create stores a new process definition.
	Create(ctx context.Context, def *ProcessDefinition) error

	// FindByKey retrieves a process definition by its unique key.
	FindByKey(ctx context.Context, key uint64) (*ProcessDefinition, error)

	// FindByProcessId retrieves the latest version of a process by its BPMN process id.
	FindLatestByProcessId(ctx context.Context, bpmnProcessId string) (*ProcessDefinition, error)

	// FindByProcessIdAndVersion retrieves a specific version.
	FindByProcessIdAndVersion(ctx context.Context, bpmnProcessId string, version uint64) (*ProcessDefinition, error)

	// FindByContentHash checks if a definition with the same content already exists.
	// Returns nil, nil if not found.
	FindByContentHash(ctx context.Context, bpmnProcessId string, hash []byte) (*ProcessDefinition, error)

	// GetLastVersion returns the highest version number for a given BPMN process id.
	// Returns 0 if no versions exist.
	GetLastVersion(ctx context.Context, bpmnProcessId string) (uint64, error)

	// Delete soft-deletes a process definition by key.
	Delete(ctx context.Context, key uint64) error
}
