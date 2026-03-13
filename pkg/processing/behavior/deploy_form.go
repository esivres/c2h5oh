package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// deployForm handles DeployFormIntent.
//
// 1. Check if form with same hash already exists (dedup).
// 2. Determine next version number.
// 3. Store form definition.
func deployForm(ctx context.Context, s storage.Store, i *intent.DeployFormIntent) ([]intent.Intent, error) {
	formRepo := s.Forms()

	// Dedup: skip if same content already deployed
	existing, err := formRepo.FindByContentHash(ctx, i.FormId, i.ContentHash)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil
	}

	// Check latest version for version numbering
	var version uint64 = 1
	latest, err := formRepo.FindLatestByFormId(ctx, i.FormId)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		version = latest.Version + 1
	}

	form := &storage.FormDefinition{
		Key:         i.Key,
		FormId:      i.FormId,
		Version:     version,
		Content:     i.Content,
		ContentHash: i.ContentHash,
		DeployedAt:  time.Now(),
	}

	return nil, formRepo.Create(ctx, form)
}
