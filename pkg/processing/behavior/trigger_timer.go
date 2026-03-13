package behavior

import (
	"context"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// triggerTimer handles TriggerTimerIntent.
//
// 1. Verify timer exists and is in Created state.
// 2. Mark timer as triggered.
// 3. Emit CompleteElementIntent for the element that owns the timer.
func triggerTimer(ctx context.Context, s storage.Store, i *intent.TriggerTimerIntent) ([]intent.Intent, error) {
	timer, err := s.Timers().GetByKey(ctx, i.TimerKey)
	if err != nil {
		return nil, err
	}
	if timer == nil {
		return nil, fmt.Errorf("timer not found: %d", i.TimerKey)
	}
	if timer.State != storage.TimerCreated {
		return nil, fmt.Errorf("timer %d is not in created state: %d", i.TimerKey, timer.State)
	}

	if err := s.Timers().Trigger(ctx, i.TimerKey); err != nil {
		return nil, err
	}

	var intents []intent.Intent

	intents = append(intents, &intent.CompleteElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: timer.ProcessInstanceKey,
		},
		ElementInstanceKey: timer.ElementInstanceKey,
	})

	// Timer cycle: schedule next repetition
	if timer.CycleDuration > 0 && timer.Repetitions != 0 {
		nextReps := timer.Repetitions
		if nextReps > 0 {
			nextReps-- // decrement remaining count
		}
		// Only create next timer if repetitions left (or infinite -1)
		if nextReps != 0 {
			intents = append(intents, &intent.CreateTimerIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: timer.ProcessInstanceKey,
				},
				ElementInstanceKey:   timer.ElementInstanceKey,
				ProcessDefinitionKey: timer.ProcessDefinitionKey,
				DueDate:              time.Now().Add(timer.CycleDuration),
				Repetitions:          nextReps,
				CycleDuration:        timer.CycleDuration,
			})
		}
	}

	return intents, nil
}
