package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedEndEvent handles ElementCompletedIntent for end events.
// Handles error/compensation/escalation/signal/terminate end events and scope completion.
func completedEndEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	bmi, err := loadBPMN(ctx, s, i.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	ei, err := s.ProcessInstances().GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, err
	}

	var intents []intent.Intent

	// Check for error end event — propagate error up the scope chain
	errorResult, err := handleErrorEndEvent(ctx, s, bmi, ei, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}
	if errorResult != nil {
		return errorResult, nil
	}

	// Check for compensation end event
	compIntents := handleCompensationThrow(bmi, ei, i.ProcessInstanceKey)
	if compIntents != nil {
		intents = append(intents, compIntents...)
	}

	// Check for escalation end event
	escResult, err := handleEscalationThrow(ctx, s, bmi, ei, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}
	if escResult != nil {
		return escResult, nil
	}

	// Check for signal end event — throw signal before completing
	signalIntents := handleSignalEndEvent(bmi, ei)
	intents = append(intents, signalIntents...)

	// Check for terminate end event
	isTerminate := isTerminateEndEvent(bmi, i.ElementId)

	if isTerminate {
		// Terminate all active element instances in the same scope
		termIntents, err := terminateScopeElements(ctx, s, i.ProcessInstanceKey, i.FlowScopeKey, i.ElementInstanceKey)
		if err != nil {
			return nil, err
		}
		intents = append(intents, termIntents...)
	}

	// Check if inside a subprocess (flowScopeKey != processInstanceKey)
	if i.FlowScopeKey != 0 && i.FlowScopeKey != ei.ProcessInstanceKey {
		// Complete the parent subprocess element
		intents = append(intents, &intent.CompleteElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey: i.FlowScopeKey,
		})
	} else {
		intents = append(intents, &intent.CompleteProcessInstanceIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
		})
	}

	return intents, nil
}
