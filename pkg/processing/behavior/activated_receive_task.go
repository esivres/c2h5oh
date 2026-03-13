package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activatedReceiveTask handles ElementActivatedIntent for receiveTask.
// Opens a message subscription to wait for a correlated message.
func activatedReceiveTask(ctx context.Context, s storage.Store, i *intent.ElementActivatedIntent) ([]intent.Intent, error) {
	if i.ProcessDefinitionKey == 0 {
		return nil, nil // wait
	}

	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN for receive task: %w", err)
	}

	fn := findFlowNode(bmi, i.ElementId)
	if fn == nil {
		return nil, nil
	}

	dom := fn.(bpmn_model.BaseElement).GetDomElement()
	if dom == nil {
		return nil, nil
	}

	// ReceiveTask can have messageRef attribute or messageEventDefinition child
	msgRef := dom.GetAttribute("messageRef")
	if msgRef == "" {
		msgDefs := dom.GetChildElementsByNS(bpmnNS, "messageEventDefinition")
		if len(msgDefs) > 0 {
			msgRef = msgDefs[0].GetAttribute("messageRef")
		}
	}

	if msgRef != "" {
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

	// No message ref — just wait
	return nil, nil
}
