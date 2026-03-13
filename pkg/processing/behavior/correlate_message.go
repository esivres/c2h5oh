package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// correlateMessage handles CorrelateMessageIntent.
//
// 1. Mark subscription as correlated.
// 2. If variables provided → emit SetVariablesIntent.
// 3. Emit CompleteElementIntent for the waiting element.
func correlateMessage(ctx context.Context, s storage.Store, i *intent.CorrelateMessageIntent) ([]intent.Intent, error) {
	msgRepo := s.MessageSubscriptions()

	if err := msgRepo.Correlate(ctx, i.SubscriptionKey); err != nil {
		return nil, err
	}

	// Look up the subscription to find the element instance
	// We need to query from all subscriptions for this process instance
	subs, err := msgRepo.FindByProcessInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}

	var elementInstanceKey uint64
	for _, sub := range subs {
		if sub.Key == i.SubscriptionKey {
			elementInstanceKey = sub.ElementInstanceKey
			break
		}
	}

	var intents []intent.Intent

	// Merge message variables into process scope
	if len(i.Variables) > 0 {
		intents = append(intents, &intent.SetVariablesIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ScopeKey:  i.ProcessInstanceKey,
			Variables: i.Variables,
		})
	}

	// Complete the element that was waiting for the message
	if elementInstanceKey != 0 {
		intents = append(intents, &intent.CompleteElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey: elementInstanceKey,
		})
	}

	return intents, nil
}
