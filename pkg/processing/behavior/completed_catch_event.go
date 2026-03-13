package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedCatchEvent handles ElementCompletedIntent for intermediate catch events.
// Takes outgoing flows and cancels event-based gateway siblings.
func completedCatchEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
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

	// Cancel event-based gateway siblings
	siblingTerminations := cancelEventBasedGatewaySiblings(ctx, s, bmi, ei, i.ProcessInstanceKey)
	intents = append(intents, siblingTerminations...)

	return intents, nil
}
