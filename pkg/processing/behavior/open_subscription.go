package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// openSubscription handles OpenSubscriptionIntent.
//
// 1. Create a message subscription.
// 2. Check if a buffered message already exists for this name + correlation key.
// 3. If found → emit CorrelateMessageIntent immediately.
func openSubscription(ctx context.Context, s storage.Store, i *intent.OpenSubscriptionIntent) ([]intent.Intent, error) {
	msgRepo := s.MessageSubscriptions()

	sub := &storage.MessageSubscription{
		Key:                i.Key,
		ProcessInstanceKey: i.ProcessInstanceKey,
		ElementInstanceKey: i.ElementInstanceKey,
		MessageName:        i.MessageName,
		CorrelationKey:     i.CorrelationKey,
		State:              storage.MessageSubscriptionOpened,
		CreatedAt:          time.Now(),
	}
	if err := msgRepo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	// Check for buffered messages
	buffered, err := msgRepo.FindBufferedMessages(ctx, i.MessageName, i.CorrelationKey)
	if err != nil {
		return nil, err
	}

	if len(buffered) > 0 {
		msg := buffered[0]
		return []intent.Intent{
			&intent.CorrelateMessageIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				SubscriptionKey:  i.Key,
				MessageBufferKey: msg.Key,
				Variables:        msg.Variables,
			},
		}, nil
	}

	return nil, nil
}
