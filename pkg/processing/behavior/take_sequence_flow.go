package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/feel"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// takeSequenceFlow handles TakeSequenceFlowIntent.
//
// 1. If condition expression is set — evaluate it against scope variables.
// 2. If condition is true (or no condition) — emit ActivateElementIntent for target.
// 3. If condition is false — no follow-up intents (flow not taken).
func takeSequenceFlow(ctx context.Context, s storage.Store, i *intent.TakeSequenceFlowIntent) ([]intent.Intent, error) {
	if i.ConditionExpression != "" {
		scope, err := loadScope(ctx, s, i.FlowScopeKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load variables for scope %d: %w", i.FlowScopeKey, err)
		}

		result, err := feel.EvalBool(i.ConditionExpression, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition %q: %w", i.ConditionExpression, err)
		}
		if !result {
			return nil, nil // condition is false, flow not taken
		}
	}

	return []intent.Intent{
		&intent.ActivateElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ProcessDefinitionKey: i.ProcessDefinitionKey,
			ElementId:            i.TargetElementId,
			FlowScopeKey:         i.FlowScopeKey,
		},
	}, nil
}
