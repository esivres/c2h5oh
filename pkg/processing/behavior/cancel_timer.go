package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// cancelTimer handles CancelTimerIntent.
//
// Marks a timer as canceled. Used when the owning element or process is terminated.
func cancelTimer(ctx context.Context, s storage.Store, i *intent.CancelTimerIntent) ([]intent.Intent, error) {
	timer, err := s.Timers().GetByKey(ctx, i.TimerKey)
	if err != nil {
		return nil, err
	}
	if timer == nil {
		return nil, fmt.Errorf("timer not found: %d", i.TimerKey)
	}

	return nil, s.Timers().Cancel(ctx, i.TimerKey)
}
