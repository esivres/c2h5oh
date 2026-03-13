package behavior

import (
	"context"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activateJob handles ActivateJobIntent.
//
// 1. Verify job exists and is in Created state.
// 2. Set worker, deadline, state=Activated.
func activateJob(ctx context.Context, s storage.Store, i *intent.ActivateJobIntent) ([]intent.Intent, error) {
	job, err := s.Jobs().GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found: %d", i.JobKey)
	}
	if job.State != storage.JobCreated {
		return nil, fmt.Errorf("job %d is not activatable, current state: %d", i.JobKey, job.State)
	}

	deadline := time.Now().Add(i.Timeout)
	return nil, s.Jobs().Activate(ctx, i.JobKey, i.Worker, deadline)
}
