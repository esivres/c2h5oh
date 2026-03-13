package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// updateJobRetries handles UpdateJobRetriesIntent.
//
// Updates the retry count for a job. If the job is in Failed state and the new
// retries are > 0, the storage layer automatically moves it back to Created state
// so it becomes activatable again.
func updateJobRetries(ctx context.Context, s storage.Store, i *intent.UpdateJobRetriesIntent) ([]intent.Intent, error) {
	job, err := s.Jobs().GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found: %d", i.JobKey)
	}
	if i.Retries < 1 {
		return nil, fmt.Errorf("retries must be > 0, got %d", i.Retries)
	}

	if err := s.Jobs().UpdateRetries(ctx, i.JobKey, i.Retries); err != nil {
		return nil, err
	}

	return nil, nil
}
