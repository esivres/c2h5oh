package behavior

import (
	"context"

	bpmn_model "github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedEventCommon extracts common logic for completed boundary/catch events:
// loads BPMN model, gets element instance, and takes outgoing sequence flows.
// Returns (nil, nil, nil, nil) if the flow node is not found.
func completedEventCommon(
	ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent,
) ([]intent.Intent, *bpmn_model.BpmnModelInstance, *storage.ElementInstance, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, nil, nil, err
	}

	ei, err := s.ProcessInstances().GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, nil, nil, err
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil, nil, nil
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

	return intents, bmi, ei, nil
}
