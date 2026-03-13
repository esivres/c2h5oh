package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// cancelProcessInstance handles CancelProcessInstanceIntent.
//
// 1. Verify process instance exists and is active.
// 2. Find all active element instances.
// 3. Emit TerminateElementIntent for each active element.
// 4. Update process instance state to Terminated.
func cancelProcessInstance(ctx context.Context, s storage.Store, i *intent.CancelProcessInstanceIntent) ([]intent.Intent, error) {
	piRepo := s.ProcessInstances()

	pi, err := piRepo.GetInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}
	if pi == nil {
		return nil, fmt.Errorf("process instance not found: %d", i.ProcessInstanceKey)
	}
	if pi.State != storage.ProcessInstanceActive {
		return nil, fmt.Errorf("process instance %d is not active, current state: %d", i.ProcessInstanceKey, pi.State)
	}

	// Find active elements
	elements, err := piRepo.FindElementInstancesByProcessInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}

	var intents []intent.Intent
	for _, ei := range elements {
		if ei.State == storage.ElementInstanceActivated || ei.State == storage.ElementInstanceActivating {
			intents = append(intents, &intent.TerminateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: ei.Key,
			})
		}
	}

	// Update process instance state
	if err := piRepo.UpdateInstanceState(ctx, i.ProcessInstanceKey, storage.ProcessInstanceTerminated); err != nil {
		return nil, err
	}

	return intents, nil
}
