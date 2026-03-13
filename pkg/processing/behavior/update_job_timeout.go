package behavior

import (
	"context"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// updateJobTimeout handles UpdateJobTimeoutIntent.
//
// Validates that the job exists and is in Activated state,
// then updates its deadline to now + timeout.
func updateJobTimeout(ctx context.Context, s storage.Store, i *intent.UpdateJobTimeoutIntent) ([]intent.Intent, error) {
	job, err := s.Jobs().GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found: %d", i.JobKey)
	}
	if job.State != storage.JobActivated {
		return nil, fmt.Errorf("job %d is not activated (state=%d), cannot update timeout", i.JobKey, job.State)
	}

	newDeadline := time.Now().Add(i.Timeout)
	return nil, s.Jobs().UpdateDeadline(ctx, i.JobKey, newDeadline)
}
