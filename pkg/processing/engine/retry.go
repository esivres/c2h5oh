package engine

import (
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
)

// RetryPolicy configures retry behavior for external intent failures.
type RetryPolicy struct {
	Intervals  []time.Duration
	MaxRetries int
}

// DefaultRetryPolicy returns a policy with 3 retries and exponential backoff.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 3,
		Intervals: []time.Duration{
			100 * time.Millisecond,
			500 * time.Millisecond,
			2 * time.Second,
		},
	}
}

func (p RetryPolicy) interval(attempt int) time.Duration {
	if attempt < len(p.Intervals) {
		return p.Intervals[attempt]
	}
	if len(p.Intervals) > 0 {
		return p.Intervals[len(p.Intervals)-1]
	}
	return time.Second
}

// retryEnvelope wraps an intent with retry metadata.
type retryEnvelope struct {
	wrapped intent.Intent
	attempt int
}

func (e *retryEnvelope) IntentType() intent.Type       { return e.wrapped.IntentType() }
func (e *retryEnvelope) GetOrigin() intent.Origin      { return e.wrapped.GetOrigin() }
func (e *retryEnvelope) GetProcessInstanceKey() uint64 { return e.wrapped.GetProcessInstanceKey() }
func (e *retryEnvelope) GetKey() uint64 {
	if kg, ok := e.wrapped.(interface{ GetKey() uint64 }); ok {
		return kg.GetKey()
	}
	return 0
}
func (e *retryEnvelope) AssignKey(k uint64) {
	if ka, ok := e.wrapped.(interface{ AssignKey(uint64) }); ok {
		ka.AssignKey(k)
	}
}

// unwrapIntent returns the original intent, stripping any retry envelope.
func unwrapIntent(i intent.Intent) intent.Intent {
	if env, ok := i.(*retryEnvelope); ok {
		return env.wrapped
	}
	return i
}
