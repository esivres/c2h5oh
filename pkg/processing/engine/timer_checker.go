package engine

import (
	"context"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
)

// TimerCheckerConfig configures the background timer checker.
type TimerCheckerConfig struct {
	// Interval is how often to check for due timers (default: 1s).
	Interval time.Duration

	// BatchSize is the max number of due timers to process per tick (default: 100).
	BatchSize int
}

// runTimerChecker periodically checks for due timers and submits TriggerTimerIntent.
// Runs until ctx is canceled.
func (p *Processor) runTimerChecker(ctx context.Context, cfg TimerCheckerConfig) {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkDueTimers(ctx, cfg.BatchSize)
		}
	}
}

func (p *Processor) checkDueTimers(ctx context.Context, batchSize int) {
	timers, err := p.store.Timers().FindDue(ctx, time.Now(), batchSize)
	if err != nil {
		p.logger.Error("failed to check due timers", "error", err)
		return
	}

	for _, timer := range timers {
		p.Submit(&intent.TriggerTimerIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: timer.ProcessInstanceKey,
			},
			TimerKey: timer.Key,
		})
	}
}
