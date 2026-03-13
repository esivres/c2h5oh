package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completedBoundaryEvent handles ElementCompletedIntent for boundary events.
// Takes outgoing flows and interrupts the attached element if cancelActivity=true.
func completedBoundaryEvent(ctx context.Context, s storage.Store, i *intent.ElementCompletedIntent) ([]intent.Intent, error) {
	intents, bmi, ei, err := completedEventCommon(ctx, s, i)
	if err != nil || bmi == nil {
		return intents, err
	}

	// Interrupt attached element if cancelActivity=true
	termIntents := handleBoundaryEventInterruption(ctx, s, bmi, ei, i.ProcessInstanceKey)
	intents = append(intents, termIntents...)

	return intents, nil
}
