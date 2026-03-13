package storage

import (
	"context"
	"time"
)

// FormDefinition represents a deployed form resource (JSON).
type FormDefinition struct {
	// Key is the unique identifier.
	Key uint64

	// FormId is the form identifier (used for lookup).
	FormId string

	// Version is the deployment version.
	Version uint64

	// Content is the raw form JSON.
	Content []byte

	// ContentHash is the hash for deduplication.
	ContentHash []byte

	// DeployedAt is the deployment timestamp.
	DeployedAt time.Time
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
