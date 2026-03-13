package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedParallelGateway handles ElementActivatedIntent for parallelGateway and inclusiveGateway.
// Implements join logic: waits for all incoming tokens before auto-completing.
func activatedParallelGateway(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return autoCompleteFromActivated(i), nil
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for parallel gateway: %w", err)
	}

	// Find the gateway element and count incoming flows
	var incomingCount int
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() == i.ElementId {
			incomingCount = len(fn.GetIncomingSequenceFlows())
			break
		}
	}

	// If 0 or 1 incoming flow, no join needed
	if incomingCount <= 1 {
		return autoCompleteFromActivated(i), nil
	}

	// Count how many activated instances exist for this gateway
	allEIs, err := s.ProcessInstances().FindElementInstancesByProcessInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}

	activatedCount := 0
	for _, ei := range allEIs {
		if ei.ElementId == i.ElementId && ei.State == storage.ElementInstanceActivated {
			activatedCount++
		}
	}

	// All tokens arrived → complete
	if activatedCount >= incomingCount {
		return autoCompleteFromActivated(i), nil
	}

	// Still waiting for more tokens — don't complete
	return nil, nil
}
