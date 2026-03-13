package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// publishMessage handles PublishMessageIntent.
//
// 1. Find open subscriptions matching name + correlation key.
// 2. If found → emit CorrelateMessageIntent for the first match.
// 3. If not found → buffer the message with TTL for later correlation.
func publishMessage(ctx context.Context, s storage.Store, i *intent.PublishMessageIntent) ([]intent.Intent, error) {
	msgRepo := s.MessageSubscriptions()

	// Try to find a matching open subscription
	subs, err := msgRepo.FindOpenSubscriptions(ctx, i.MessageName, i.CorrelationKey)
	if err != nil {
		return nil, err
	}

	if len(subs) > 0 {
		// Correlate with the first matching subscription
		sub := subs[0]
		return []intent.Intent{
			&intent.CorrelateMessageIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: sub.ProcessInstanceKey,
				},
				SubscriptionKey: sub.Key,
				Variables:       i.Variables,
			},
		}, nil
	}

	// No subscription found — buffer the message
	ttl := i.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute // default TTL
	}

	msg := &storage.MessageBuffer{
		Key:            i.Key,
		MessageName:    i.MessageName,
		CorrelationKey: i.CorrelationKey,
		Variables:      i.Variables,
		ExpiresAt:      time.Now().Add(ttl),
		CreatedAt:      time.Now(),
	}
	return nil, msgRepo.BufferMessage(ctx, msg)
}
