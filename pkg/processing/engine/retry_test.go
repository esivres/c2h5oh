package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry_ExternalIntent_RetriesBeforeIncident(t *testing.T) {
	store := newMockStore()

	var callCount atomic.Int32

	registry := behavior.NewRegistry()
	registry.Register(intent.DeployProcess, behavior.Typed(
		func(_ context.Context, _ storage.Store, _ *intent.DeployProcessIntent) ([]intent.Intent, error) {
			callCount.Add(1)
			return nil, fmt.Errorf("transient error")
		},
	))
	registry.Register(intent.CreateIncident, behavior.Typed(
		func(_ context.Context, s storage.Store, i *intent.CreateIncidentIntent) ([]intent.Intent, error) {
			return nil, s.Incidents().Create(context.Background(), &storage.Incident{
				Key:          i.Key,
				Type:         storage.IncidentType(i.ErrorType),
				ErrorMessage: i.ErrorMessage,
			})
		},
	))

	fastPolicy := RetryPolicy{
		MaxRetries: 3,
		Intervals:  []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
	}

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
		RetryPolicy: &fastPolicy,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test",
		Content:       []byte("<test/>"),
	})

	// Wait for all retries + final incident
	require.Eventually(t, func() bool {
		return callCount.Load() >= 4 // 1 initial + 3 retries
	}, 2*time.Second, 10*time.Millisecond)

	// Should have created an incident after retries exhausted
	require.Eventually(t, func() bool {
		store.incidents.mu.Lock()
		defer store.incidents.mu.Unlock()
		return len(store.incidents.created) >= 1
	}, 2*time.Second, 10*time.Millisecond)

	store.incidents.mu.Lock()
	defer store.incidents.mu.Unlock()
	assert.Equal(t, "INTENT_PROCESSING", string(store.incidents.created[0].Type))
}

func TestRetry_ExternalIntent_SucceedsOnRetry(t *testing.T) {
	store := newMockStore()

	var callCount atomic.Int32

	registry := behavior.NewRegistry()
	registry.Register(intent.DeployProcess, behavior.Typed(
		func(_ context.Context, s storage.Store, i *intent.DeployProcessIntent) ([]intent.Intent, error) {
			n := callCount.Add(1)
			if n < 3 {
				return nil, fmt.Errorf("transient error")
			}
			// Succeed on 3rd attempt
			_ = s.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
				Key:           i.Key,
				BpmnProcessId: i.BpmnProcessId,
				Version:       1,
				Content:       i.Content,
			})
			return nil, nil
		},
	))

	fastPolicy := RetryPolicy{
		MaxRetries: 3,
		Intervals:  []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
	}

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
		RetryPolicy: &fastPolicy,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test",
		Content:       []byte("<test/>"),
	})

	// Should succeed on 3rd attempt
	require.Eventually(t, func() bool {
		return callCount.Load() >= 3
	}, 2*time.Second, 10*time.Millisecond)

	// Give time for any incident processing
	time.Sleep(50 * time.Millisecond)

	// No incidents should have been created
	store.incidents.mu.Lock()
	defer store.incidents.mu.Unlock()
	assert.Empty(t, store.incidents.created, "no incident should be created when retry succeeds")
}

func TestRetry_InternalIntent_NoRetry(t *testing.T) {
	store := newMockStore()

	var callCount atomic.Int32
	var mu sync.Mutex

	registry := behavior.NewRegistry()
	registry.Register(intent.CompleteProcessInstance, behavior.Typed(
		func(_ context.Context, _ storage.Store, _ *intent.CompleteProcessInstanceIntent) ([]intent.Intent, error) {
			callCount.Add(1)
			return nil, fmt.Errorf("internal error")
		},
	))
	// Need CreateIncident registered for the incident to be processed
	registry.Register(intent.CreateIncident, behavior.Typed(
		func(_ context.Context, _ storage.Store, i *intent.CreateIncidentIntent) ([]intent.Intent, error) {
			mu.Lock()
			defer mu.Unlock()
			return nil, nil
		},
	))

	fastPolicy := RetryPolicy{
		MaxRetries: 3,
		Intervals:  []time.Duration{10 * time.Millisecond},
	}

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
		RetryPolicy: &fastPolicy,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.CompleteProcessInstanceIntent{
		Header: intent.Header{Origin: intent.Internal, ProcessInstanceKey: 100},
	})

	// Wait for processing
	require.Eventually(t, func() bool {
		return callCount.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond)

	// Give time for any retry
	time.Sleep(100 * time.Millisecond)

	// Internal intent should NOT retry — only called once
	assert.Equal(t, int32(1), callCount.Load(), "internal intent should not be retried")
}
