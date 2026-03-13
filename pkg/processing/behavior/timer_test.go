package behavior

import (
	"context"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func seedTimer(t *testing.T, s storage.Store, key, piKey, eiKey uint64, dueDate time.Time, state storage.TimerState) {
	t.Helper()
	require.NoError(t, s.Timers().Create(context.Background(), &storage.Timer{
		Key:                key,
		ProcessInstanceKey: piKey,
		ElementInstanceKey: eiKey,
		State:              state,
		DueDate:            dueDate,
		Repetitions:        0,
		CreatedAt:          time.Now(),
	}))
	if state == storage.TimerTriggered {
		require.NoError(t, s.Timers().Trigger(context.Background(), key))
	}
	if state == storage.TimerCanceled {
		require.NoError(t, s.Timers().Cancel(context.Background(), key))
	}
}

func TestCreateTimer_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(createTimer),
		&intent.CreateTimerIntent{
			Header:               intent.Header{Key: 1, ProcessInstanceKey: 100},
			ElementInstanceKey:   200,
			ProcessDefinitionKey: 1,
			DueDate:              time.Now().Add(time.Hour),
			Repetitions:          0,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	timer, err := store.Timers().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, timer)
	assert.Equal(t, storage.TimerCreated, timer.State)
	assert.Equal(t, uint64(200), timer.ElementInstanceKey)
}

func TestTriggerTimer_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedTimer(t, store, 1, 100, 200, time.Now().Add(-time.Minute), storage.TimerCreated)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(triggerTimer),
		&intent.TriggerTimerIntent{
			Header:   intent.Header{Key: 10, ProcessInstanceKey: 100},
			TimerKey: 1,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)

	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(200), complete.ElementInstanceKey)

	timer, err := store.Timers().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.TimerTriggered, timer.State)
}

func TestTriggerTimer_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(triggerTimer),
		&intent.TriggerTimerIntent{
			Header:   intent.Header{Key: 10, ProcessInstanceKey: 100},
			TimerKey: 999,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timer not found")
}

func TestTriggerTimer_AlreadyTriggered(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedTimer(t, store, 1, 100, 200, time.Now().Add(-time.Minute), storage.TimerTriggered)

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(triggerTimer),
		&intent.TriggerTimerIntent{
			Header:   intent.Header{Key: 10, ProcessInstanceKey: 100},
			TimerKey: 1,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in created state")
}

func TestCancelTimer_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedTimer(t, store, 1, 100, 200, time.Now().Add(time.Hour), storage.TimerCreated)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(cancelTimer),
		&intent.CancelTimerIntent{
			Header:   intent.Header{Key: 10, ProcessInstanceKey: 100},
			TimerKey: 1,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	timer, err := store.Timers().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.TimerCanceled, timer.State)
}

func TestCancelTimer_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(cancelTimer),
		&intent.CancelTimerIntent{
			Header:   intent.Header{Key: 10, ProcessInstanceKey: 100},
			TimerKey: 999,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timer not found")
}

func TestFindDueTimers(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Past timer — should be found
	seedTimer(t, store, 1, 100, 200, time.Now().Add(-time.Minute), storage.TimerCreated)
	// Future timer — should not be found
	seedTimer(t, store, 2, 100, 201, time.Now().Add(time.Hour), storage.TimerCreated)

	due, err := store.Timers().FindDue(context.Background(), time.Now(), 10)
	require.NoError(t, err)
	require.Len(t, due, 1)
	assert.Equal(t, uint64(1), due[0].Key)
}
