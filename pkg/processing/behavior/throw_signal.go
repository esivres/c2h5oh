package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// throwSignal handles ThrowSignalIntent.
//
// Broadcasts a signal to all open subscriptions matching the signal name.
// Signals use message subscriptions with empty correlation key.
func throwSignal(ctx context.Context, s storage.Store, i *intent.ThrowSignalIntent) ([]intent.Intent, error) {
	// Find all open subscriptions for this signal name (correlation key = "")
	subs, err := s.MessageSubscriptions().FindOpenSubscriptions(ctx, i.SignalName, "")
	if err != nil {
		return nil, err
	}

	var intents []intent.Intent
	for _, sub := range subs {
		intents = append(intents, &intent.CorrelateMessageIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: sub.ProcessInstanceKey,
			},
			SubscriptionKey: sub.Key,
			Variables:       i.Variables,
		})
	}

	return intents, nil
}
