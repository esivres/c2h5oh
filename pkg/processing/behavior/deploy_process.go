package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// deployProcess handles DeployProcessIntent.
//
// 1. Check if a definition with the same content hash already exists (idempotent).
// 2. If exists — return with no follow-up intents (already deployed).
// 3. Get last version, increment, store new definition.
// 4. No follow-up intents.
func deployProcess(ctx context.Context, s storage.Store, i *intent.DeployProcessIntent) ([]intent.Intent, error) {
	repo := s.ProcessDefinitions()

	// Idempotent: same content hash → already deployed
	existing, err := repo.FindByContentHash(ctx, i.BpmnProcessId, i.ContentHash)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	// Get next version
	lastVersion, err := repo.GetLastVersion(ctx, i.BpmnProcessId)
	if err != nil {
		return nil, err
	}
	nextVersion := lastVersion + 1

	def := &storage.ProcessDefinition{
		Key:           i.Key,
		BpmnProcessId: i.BpmnProcessId,
		Name:          i.Name,
		Version:       nextVersion,
		ContentHash:   i.ContentHash,
		Content:       i.Content,
		DeployedAt:    time.Now(),
	}

	if err := repo.Create(ctx, def); err != nil {
		return nil, err
	}

	return nil, nil
}
