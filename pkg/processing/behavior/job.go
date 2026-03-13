package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// createJob handles CreateJobIntent.
func createJob(ctx context.Context, s storage.Store, i *intent.CreateJobIntent) ([]intent.Intent, error) {
	job := &storage.Job{
		Key:                  i.Key,
		ProcessInstanceKey:   i.ProcessInstanceKey,
		ElementInstanceKey:   i.ElementInstanceKey,
		ProcessDefinitionKey: i.ProcessDefinitionKey,
		Type:                 i.Type,
		State:                storage.JobCreated,
		Retries:              i.Retries,
		Variables:            i.Variables,
		CreatedAt:            time.Now(),
	}
	return nil, s.Jobs().Create(ctx, job)
}

// completeJob handles CompleteJobIntent.
//
// 1. Mark job as completed.
// 2. Emit CompleteElementIntent for the element that owns the job.
func completeJob(ctx context.Context, s storage.Store, i *intent.CompleteJobIntent) ([]intent.Intent, error) {
	jobRepo := s.Jobs()

	job, err := jobRepo.GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}

	if err := jobRepo.Complete(ctx, i.JobKey, i.Variables); err != nil {
		return nil, err
	}

	return []intent.Intent{
		&intent.CompleteElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey: job.ElementInstanceKey,
			Variables:          i.Variables,
		},
	}, nil
}

// failJob handles FailJobIntent.
//
// 1. Update job retries and state.
// 2. If retries exhausted — emit CreateIncidentIntent.
func failJob(ctx context.Context, s storage.Store, i *intent.FailJobIntent) ([]intent.Intent, error) {
	if err := s.Jobs().Fail(ctx, i.JobKey, i.Retries, i.ErrorMessage); err != nil {
		return nil, err
	}

	if i.Retries <= 0 {
		job, err := s.Jobs().GetByKey(ctx, i.JobKey)
		if err != nil {
			return nil, err
		}
		return []intent.Intent{
			&intent.CreateIncidentIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: job.ElementInstanceKey,
				JobKey:             i.JobKey,
				ErrorType:          "JOB_NO_RETRIES",
				ErrorMessage:       i.ErrorMessage,
			},
		}, nil
	}

	return nil, nil
}
