package behavior

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func TestPublishMessage_WithMatchingSubscription(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Create an open subscription
	require.NoError(t, store.MessageSubscriptions().CreateSubscription(context.Background(), &storage.MessageSubscription{
		Key:                1,
		ProcessInstanceKey: 100,
		ElementInstanceKey: 200,
		MessageName:        "payment-received",
		CorrelationKey:     "ORD-123",
		State:              storage.MessageSubscriptionOpened,
		CreatedAt:          time.Now(),
	}))

	vars, _ := json.Marshal(map[string]any{"paid": true})

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(publishMessage),
		&intent.PublishMessageIntent{
			Header:         intent.Header{Key: 10},
			MessageName:    "payment-received",
			CorrelationKey: "ORD-123",
			Variables:      vars,
			TTL:            time.Minute,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)

	corr, ok := intents[0].(*intent.CorrelateMessageIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(1), corr.SubscriptionKey)
	assert.Equal(t, uint64(100), corr.ProcessInstanceKey)
	assert.Equal(t, vars, corr.Variables)
}

func TestPublishMessage_NoSubscription_BuffersMessage(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	vars, _ := json.Marshal(map[string]any{"data": "test"})

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(publishMessage),
		&intent.PublishMessageIntent{
			Header:         intent.Header{Key: 10},
			MessageName:    "order-created",
			CorrelationKey: "ORD-456",
			Variables:      vars,
			TTL:            5 * time.Minute,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	// Message should be buffered
	buffered, err := store.MessageSubscriptions().FindBufferedMessages(context.Background(), "order-created", "ORD-456")
	require.NoError(t, err)
	require.Len(t, buffered, 1)
	assert.Equal(t, vars, buffered[0].Variables)
}

func TestOpenSubscription_NoBufferedMessage(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(openSubscription),
		&intent.OpenSubscriptionIntent{
			Header:             intent.Header{Key: 1, ProcessInstanceKey: 100},
			ElementInstanceKey: 200,
			MessageName:        "order-shipped",
			CorrelationKey:     "ORD-789",
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents) // No buffered message, just wait

	// Subscription should exist
	subs, err := store.MessageSubscriptions().FindOpenSubscriptions(context.Background(), "order-shipped", "ORD-789")
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, storage.MessageSubscriptionOpened, subs[0].State)
}

func TestOpenSubscription_WithBufferedMessage(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Buffer a message first
	vars, _ := json.Marshal(map[string]any{"shipped": true})
	require.NoError(t, store.MessageSubscriptions().BufferMessage(context.Background(), &storage.MessageBuffer{
		Key:            50,
		MessageName:    "order-shipped",
		CorrelationKey: "ORD-789",
		Variables:      vars,
		ExpiresAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now(),
	}))

	// Now open subscription — should immediately correlate
	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(openSubscription),
		&intent.OpenSubscriptionIntent{
			Header:             intent.Header{Key: 1, ProcessInstanceKey: 100},
			ElementInstanceKey: 200,
			MessageName:        "order-shipped",
			CorrelationKey:     "ORD-789",
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)

	corr, ok := intents[0].(*intent.CorrelateMessageIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(1), corr.SubscriptionKey)
	assert.Equal(t, uint64(50), corr.MessageBufferKey)
	assert.Equal(t, vars, corr.Variables)
}

func TestCorrelateMessage_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Create subscription
	require.NoError(t, store.MessageSubscriptions().CreateSubscription(context.Background(), &storage.MessageSubscription{
		Key:                1,
		ProcessInstanceKey: 100,
		ElementInstanceKey: 200,
		MessageName:        "test-msg",
		CorrelationKey:     "KEY-1",
		State:              storage.MessageSubscriptionOpened,
		CreatedAt:          time.Now(),
	}))

	vars, _ := json.Marshal(map[string]any{"result": "ok"})

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(correlateMessage),
		&intent.CorrelateMessageIntent{
			Header:          intent.Header{Key: 10, ProcessInstanceKey: 100},
			SubscriptionKey: 1,
			Variables:       vars,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 2) // SetVariables + CompleteElement

	setVars, ok := intents[0].(*intent.SetVariablesIntent)
	require.True(t, ok)
	assert.Equal(t, vars, setVars.Variables)

	complete, ok := intents[1].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(200), complete.ElementInstanceKey)
}

func TestCorrelateMessage_NoVariables(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	require.NoError(t, store.MessageSubscriptions().CreateSubscription(context.Background(), &storage.MessageSubscription{
		Key:                1,
		ProcessInstanceKey: 100,
		ElementInstanceKey: 200,
		MessageName:        "test-msg",
		CorrelationKey:     "KEY-1",
		State:              storage.MessageSubscriptionOpened,
		CreatedAt:          time.Now(),
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(correlateMessage),
		&intent.CorrelateMessageIntent{
			Header:          intent.Header{Key: 10, ProcessInstanceKey: 100},
			SubscriptionKey: 1,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1) // Only CompleteElement, no SetVariables
}

func TestCloseSubscription_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	require.NoError(t, store.MessageSubscriptions().CreateSubscription(context.Background(), &storage.MessageSubscription{
		Key:                1,
		ProcessInstanceKey: 100,
		ElementInstanceKey: 200,
		MessageName:        "test-msg",
		CorrelationKey:     "KEY-1",
		State:              storage.MessageSubscriptionOpened,
		CreatedAt:          time.Now(),
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(closeSubscription),
		&intent.CloseSubscriptionIntent{
			Header:          intent.Header{Key: 10, ProcessInstanceKey: 100},
			SubscriptionKey: 1,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	// Should be closed
	subs, err := store.MessageSubscriptions().FindOpenSubscriptions(context.Background(), "test-msg", "KEY-1")
	require.NoError(t, err)
	assert.Empty(t, subs)
}

func TestCloseSubscription_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(closeSubscription),
		&intent.CloseSubscriptionIntent{
			Header:          intent.Header{Key: 10, ProcessInstanceKey: 100},
			SubscriptionKey: 999,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscription not found")
}
