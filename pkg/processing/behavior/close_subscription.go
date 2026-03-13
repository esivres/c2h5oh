package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// closeSubscription handles CloseSubscriptionIntent.
//
// Marks a subscription as closed. Used when the element is completed or terminated.
func closeSubscription(ctx context.Context, s storage.Store, i *intent.CloseSubscriptionIntent) ([]intent.Intent, error) {
	// Verify subscription exists by looking at process instance subscriptions
	subs, err := s.MessageSubscriptions().FindByProcessInstance(ctx, i.ProcessInstanceKey)
	if err != nil {
		return nil, err
	}

	found := false
	for _, sub := range subs {
		if sub.Key == i.SubscriptionKey {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("subscription not found: %d", i.SubscriptionKey)
	}

	return nil, s.MessageSubscriptions().CloseSubscription(ctx, i.SubscriptionKey)
}
