package storage

import (
	"context"
	"time"
)

// FormDefinition represents a deployed form resource (JSON).
type FormDefinition struct {
	DeployedAt  time.Time
	FormId      string
	Content     []byte
	ContentHash []byte
	Key         uint64
	Version     uint64
}

// FormRepository manages form definition storage.
type FormRepository interface {
	// Create stores a new form definition.
	Create(ctx context.Context, form *FormDefinition) error

	// FindByKey retrieves a form by key.
	FindByKey(ctx context.Context, key uint64) (*FormDefinition, error)

	// FindLatestByFormId finds the latest version of a form by its id.
	FindLatestByFormId(ctx context.Context, formId string) (*FormDefinition, error)

	// FindByContentHash finds a form by formId and content hash (for dedup).
	FindByContentHash(ctx context.Context, formId string, hash []byte) (*FormDefinition, error)

	// Delete soft-deletes a form definition by key.
	Delete(ctx context.Context, key uint64) error
}
