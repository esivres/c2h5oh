package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedCatchEvent handles ElementActivatedIntent for intermediateCatchEvent and boundaryEvent.
// Detects the event definition type (timer, message, signal, link) and creates the appropriate intent.
// The element does NOT auto-complete — it waits for the event to trigger.
func activatedCatchEvent(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return nil, nil // wait (no model to inspect)
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for catch event: %w", err)
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	dom := fn.(bpmn_model.BaseElement).GetDomElement()
	if dom == nil {
		return nil, nil
	}

	// Check for TimerEventDefinition
	timerDefs := dom.GetChildElementsByNS(bpmnNS, "timerEventDefinition")
	if len(timerDefs) > 0 {
		spec, err := parseTimerSpec(timerDefs[0], bpmnNS)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timer on %q: %w", i.ElementId, err)
		}
		return []intent.Intent{
			&intent.CreateTimerIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey:   i.ElementInstanceKey,
				ProcessDefinitionKey: i.ProcessDefinitionKey,
				DueDate:              spec.DueDate,
				Repetitions:          spec.Repetitions,
				CycleDuration:        spec.CycleDuration,
			},
		}, nil
	}

	// Check for MessageEventDefinition
	msgDefs := dom.GetChildElementsByNS(bpmnNS, "messageEventDefinition")
	if len(msgDefs) > 0 {
		msgRef := msgDefs[0].GetAttribute("messageRef")
		msgName := resolveMessageName(bmi, msgRef)
		corrKey := resolveCorrelationKey(ctx, s, bmi, msgRef, i.FlowScopeKey)

		return []intent.Intent{
			&intent.OpenSubscriptionIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.ElementInstanceKey,
				MessageName:        msgName,
				CorrelationKey:     corrKey,
			},
		}, nil
	}

	// Check for LinkEventDefinition (catch side — auto-complete)
	linkDefs := dom.GetChildElementsByNS(bpmnNS, "linkEventDefinition")
	if len(linkDefs) > 0 {
		return autoCompleteFromActivated(i), nil
	}

	// Check for SignalEventDefinition (catch)
	sigDefs := dom.GetChildElementsByNS(bpmnNS, "signalEventDefinition")
	if len(sigDefs) > 0 {
		sigRef := sigDefs[0].GetAttribute("signalRef")
		sigName := resolveSignalName(bmi, sigRef)

		return []intent.Intent{
			&intent.OpenSubscriptionIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.ElementInstanceKey,
				MessageName:        sigName,
				CorrelationKey:     "", // signals are broadcast
			},
		}, nil
	}

	// Unknown event definition type — just wait
	return nil, nil
}
