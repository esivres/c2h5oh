package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// terminateElement handles TerminateElementIntent.
//
// 1. Update element instance state to Terminated.
// 2. If the element has an active job → fail the job.
func terminateElement(ctx context.Context, s storage.Store, i *intent.TerminateElementIntent) ([]intent.Intent, error) {
	piRepo := s.ProcessInstances()

	ei, err := piRepo.GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, err
	}
	if ei == nil {
		return nil, fmt.Errorf("element instance not found: %d", i.ElementInstanceKey)
	}

	if err := piRepo.UpdateElementInstanceState(ctx, i.ElementInstanceKey, storage.ElementInstanceTerminated); err != nil {
		return nil, err
	}

	// If this is a service task, cancel any active jobs
	if ei.ElementType == "serviceTask" {
		jobs, err := s.Jobs().FindByProcessInstance(ctx, ei.ProcessInstanceKey)
		if err != nil {
			return nil, err
		}
		for _, job := range jobs {
			if job.ElementInstanceKey == i.ElementInstanceKey &&
				(job.State == storage.JobCreated || job.State == storage.JobActivated) {
				if err := s.Jobs().Fail(ctx, job.Key, 0, "element terminated"); err != nil {
					return nil, err
				}
			}
		}
	}

	return nil, nil
}
