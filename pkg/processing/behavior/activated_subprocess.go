package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedSubProcess handles ElementActivatedIntent for subProcess.
// Finds the inner start event and activates it with flowScopeKey = subprocess instance key.
func activatedSubProcess(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return autoCompleteFromActivated(i), nil
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for subprocess: %w", err)
	}

	subProcesses := bpmn_model.GetTypedElements[bpmn_model.SubProcess](bmi.ModelInstance)
	for _, sp := range subProcesses {
		if sp.GetId() != i.ElementId {
			continue
		}

		for _, fe := range sp.GetFlowElements() {
			if _, ok := fe.(bpmn_model.StartEvent); ok {
				return []intent.Intent{
					&intent.ActivateElementIntent{
						Header: intent.Header{
							Origin:             intent.Internal,
							ProcessInstanceKey: i.ProcessInstanceKey,
						},
						ProcessDefinitionKey: i.ProcessDefinitionKey,
						ElementId:            fe.GetId(),
						ElementType:          "startEvent",
						FlowScopeKey:         i.ElementInstanceKey, // subprocess instance is the scope
					},
				}, nil
			}
		}
		break
	}

	return autoCompleteFromActivated(i), nil
}
