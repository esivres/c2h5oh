package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// deleteResource handles DeleteResourceIntent.
// Soft-deletes a process definition or form by key.
func deleteResource(ctx context.Context, s storage.Store, i *intent.DeleteResourceIntent) ([]intent.Intent, error) {
	// Try as process definition
	def, err := s.ProcessDefinitions().FindByKey(ctx, i.ResourceKey)
	if err != nil {
		return nil, err
	}
	if def != nil {
		return nil, s.ProcessDefinitions().Delete(ctx, i.ResourceKey)
	}

	// Try as form
	form, err := s.Forms().FindByKey(ctx, i.ResourceKey)
	if err != nil {
		return nil, err
	}
	if form != nil {
		return nil, s.Forms().Delete(ctx, i.ResourceKey)
	}

	return nil, fmt.Errorf("resource not found: %d", i.ResourceKey)
}
