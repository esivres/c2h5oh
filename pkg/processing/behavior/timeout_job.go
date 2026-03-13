package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// timeoutJob handles TimeOutJobIntent.
//
// 1. Verify job exists and is in Activated state.
// 2. Decrement retries.
// 3. If retries > 0 → return job to Created state (available for re-activation).
// 4. If retries <= 0 → FailJob (which will create an incident).
func timeoutJob(ctx context.Context, s storage.Store, i *intent.TimeOutJobIntent) ([]intent.Intent, error) {
	job, err := s.Jobs().GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found: %d", i.JobKey)
	}
	if job.State != storage.JobActivated {
		return nil, fmt.Errorf("job %d is not activated, current state: %d", i.JobKey, job.State)
	}

	newRetries := job.Retries - 1
	if newRetries > 0 {
		// Return to Created state — worker can pick it up again
		return nil, s.Jobs().Fail(ctx, i.JobKey, newRetries, "job timed out")
	}

	// No retries left — fail the job, which will create an incident
	return []intent.Intent{
		&intent.FailJobIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			JobKey:       i.JobKey,
			Retries:      0,
			ErrorMessage: "job timed out, no retries left",
		},
	}, nil
}
