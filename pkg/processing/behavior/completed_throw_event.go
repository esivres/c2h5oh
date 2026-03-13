package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedThrowEvent handles ElementCompletedIntent for intermediate throw events.
// Takes outgoing flows and handles compensation/link/escalation throw semantics.
func completedThrowEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
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

	// Compensation throw: activate compensation handlers
	compIntents := handleCompensationThrow(bmi, ei, i.ProcessInstanceKey)
	intents = append(intents, compIntents...)

	// Link throw: teleport to matching catch link event (only if no outgoing flows)
	if len(outgoing) == 0 {
		linkIntents := handleLinkThrowEvent(bmi, ei, i.ProcessInstanceKey)
		if linkIntents != nil {
			return linkIntents, nil
		}
	}

	// Escalation throw: propagate escalation to parent scope
	escIntents, err := handleEscalationThrow(ctx, s, bmi, ei, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}
	// Filter out incidents for intermediate throw — escalation is optional
	for _, in := range escIntents {
		if _, isIncident := in.(*intent.CreateIncidentIntent); !isIncident {
			intents = append(intents, in)
		}
	}

	return intents, nil
}
