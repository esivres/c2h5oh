package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedAutoComplete is the default fallback for ElementActivatedIntent.
// It simply emits a CompleteElementIntent (auto-complete).
// Used for: startEvent, endEvent, manualTask, eventBasedGateway, and unknown types.
func activatedAutoComplete(_ context.Context, _ storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	return autoCompleteFromActivated(i), nil
}
