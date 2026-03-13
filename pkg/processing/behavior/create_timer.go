package behavior

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// createTimer handles CreateTimerIntent.
//
// Stores a new timer in the repository. The timer checker goroutine
// will later find due timers and emit TriggerTimerIntent.
func createTimer(ctx context.Context, s storage.Store, i *intent.CreateTimerIntent) ([]intent.Intent, error) {
	timer := &storage.Timer{
		Key:                  i.Key,
		ProcessInstanceKey:   i.ProcessInstanceKey,
		ElementInstanceKey:   i.ElementInstanceKey,
		ProcessDefinitionKey: i.ProcessDefinitionKey,
		State:                storage.TimerCreated,
		DueDate:              i.DueDate,
		Repetitions:          i.Repetitions,
		CycleDuration:        i.CycleDuration,
		CreatedAt:            time.Now(),
	}
	return nil, s.Timers().Create(ctx, timer)
}
