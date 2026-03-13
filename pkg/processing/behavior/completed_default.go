package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedDefault is the default fallback for ElementCompletedIntent.
// It takes all outgoing sequence flows and activates their targets.
// If no outgoing flows and not an end event, checks ad-hoc subprocess completion.
func completedDefault(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	var intents []intent.Intent
	outgoing := fn.GetOutgoingSequenceFlows()
	for _, sf := range outgoing {
		target := sf.GetTarget()
		if target == nil {
			continue
		}
		intents = append(intents, activateTarget(i.ProcessInstanceKey, i.ProcessDefinitionKey, i.FlowScopeKey, target, bmi))
	}

	// Non-endEvent element with no outgoing flows inside an ad-hoc subprocess
	// → check if all ad-hoc children are done
	if len(intents) == 0 {
		ei, err := s.ProcessInstances().GetElementInstance(ctx, i.ElementInstanceKey)
		if err != nil {
			return nil, err
		}
		if ei != nil {
			adHocIntents, err := checkAdHocSubProcessCompletion(ctx, s, ei, i.ProcessInstanceKey)
			if err != nil {
				return nil, err
			}
			if adHocIntents != nil {
				return adHocIntents, nil
			}
		}
	}

	return intents, nil
}
