package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedThrowEvent handles ElementActivatedIntent for intermediateThrowEvent.
// Detects signal throw events and emits ThrowSignalIntent, otherwise auto-completes.
func activatedThrowEvent(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return autoCompleteFromActivated(i), nil
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for throw event: %w", err)
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return autoCompleteFromActivated(i), nil
	}

	dom := fn.(bpmn_model.BaseElement).GetDomElement()
	if dom == nil {
		return autoCompleteFromActivated(i), nil
	}

	sigDefs := dom.GetChildElementsByNS(bpmnNS, "signalEventDefinition")
	if len(sigDefs) > 0 {
		sigRef := sigDefs[0].GetAttribute("signalRef")
		sigName := resolveSignalName(bmi, sigRef)

		// Throw signal + auto-complete
		return []intent.Intent{
			&intent.ThrowSignalIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				SignalName: sigName,
			},
			&intent.CompleteElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.ElementInstanceKey,
			},
		}, nil
	}

	return autoCompleteFromActivated(i), nil
}
