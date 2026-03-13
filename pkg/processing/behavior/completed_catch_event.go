package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedCatchEvent handles ElementCompletedIntent for intermediate catch events.
// Takes outgoing flows and cancels event-based gateway siblings.
func completedCatchEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	intents, bmi, _, err := completedEventCommon(ctx, s, i)
	if err != nil || bmi == nil {
		return intents, err
	}

	// Cancel event-based gateway siblings
	ei, err := s.ProcessInstances().GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, err
	}
	siblingTerminations := cancelEventBasedGatewaySiblings(ctx, s, bmi, ei, i.ProcessInstanceKey)
	intents = append(intents, siblingTerminations...)

	return intents, nil
}
