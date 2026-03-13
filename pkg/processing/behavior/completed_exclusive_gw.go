package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedExclusiveGateway handles ElementCompletedIntent for exclusive gateways.
// Evaluates conditions on outgoing flows and takes the first true (or default).
func completedExclusiveGateway(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	gw, ok := fn.(bpmn_model.ExclusiveGateway)
	if !ok {
		return nil, nil
	}

	selected, err := resolveExclusiveGateway(ctx, s, gw, i.FlowScopeKey)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return []intent.Intent{
			&intent.CreateIncidentIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.ElementInstanceKey,
				ErrorType:          "CONDITION_ERROR",
				ErrorMessage:       fmt.Sprintf("no outgoing flow taken from exclusive gateway %q", i.ElementId),
			},
		}, nil
	}

	return []intent.Intent{
		activateTarget(i.ProcessInstanceKey, i.ProcessDefinitionKey, i.FlowScopeKey, selected, bmi),
	}, nil
}
