package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedBoundaryEvent handles ElementCompletedIntent for boundary events.
// Takes outgoing flows and interrupts the attached element if cancelActivity=true.
func completedBoundaryEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	ei, err := s.ProcessInstances().GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, err
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	// Take outgoing flows
	var intents []intent.Intent
	outgoing := fn.GetOutgoingSequenceFlows()
	for _, sf := range outgoing {
		target := sf.GetTarget()
		if target == nil {
			continue
		}
		intents = append(intents, activateTarget(i.ProcessInstanceKey, i.ProcessDefinitionKey, i.FlowScopeKey, target, bmi))
	}

	// Interrupt attached element if cancelActivity=true
	termIntents := handleBoundaryEventInterruption(ctx, s, bmi, ei, i.ProcessInstanceKey)
	intents = append(intents, termIntents...)

	return intents, nil
}
