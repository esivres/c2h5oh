package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedInclusiveGateway handles ElementCompletedIntent for inclusive gateways.
// Evaluates conditions on all outgoing flows and takes all that are true (or default).
func completedInclusiveGateway(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	gw, ok := fn.(bpmn_model.InclusiveGateway)
	if !ok {
		return nil, nil
	}

	targets, err := resolveInclusiveGateway(ctx, s, gw, i.FlowScopeKey)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return []intent.Intent{
			&intent.CreateIncidentIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.ElementInstanceKey,
				ErrorType:          "CONDITION_ERROR",
				ErrorMessage:       fmt.Sprintf("no outgoing flow taken from inclusive gateway %q", i.ElementId),
			},
		}, nil
	}

	var intents []intent.Intent
	for _, target := range targets {
		intents = append(intents, activateTarget(i.ProcessInstanceKey, i.ProcessDefinitionKey, i.FlowScopeKey, target, bmi))
	}
	return intents, nil
}
