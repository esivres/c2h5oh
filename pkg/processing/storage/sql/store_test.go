package sql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, Migrate(context.Background(), db))
	return db
}

// --- ProcessDefinition ---

func TestProcessDefinition_CreateAndFind(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	hash := sha256.Sum256([]byte("<xml>test</xml>"))
	def := &storage.ProcessDefinition{
		Key:           1001,
		BpmnProcessId: "order-process",
		Name:          "Order Process",
		Version:       1,
		ContentHash:   hash[:],
		Content:       []byte("<xml>test</xml>"),
		DeployedAt:    time.Now().Truncate(time.Millisecond),
	}

	err := store.ProcessDefinitions().Create(ctx, def)
	require.NoError(t, err)

	// FindByKey
	found, err := store.ProcessDefinitions().FindByKey(ctx, 1001)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "order-process", found.BpmnProcessId)
	assert.Equal(t, uint64(1), found.Version)

	// FindLatestByProcessId
	found, err = store.ProcessDefinitions().FindLatestByProcessId(ctx, "order-process")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, uint64(1), found.Version)

	// FindByContentHash
	found, err = store.ProcessDefinitions().FindByContentHash(ctx, "order-process", hash[:])
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, uint64(1001), found.Key)

	// GetLastVersion
	ver, err := store.ProcessDefinitions().GetLastVersion(ctx, "order-process")
	require.NoError(t, err)
	assert.Equal(t, uint64(1), ver)

	// GetLastVersion — unknown process
	ver, err = store.ProcessDefinitions().GetLastVersion(ctx, "unknown")
	require.NoError(t, err)
	assert.Equal(t, uint64(0), ver)
}

func TestProcessDefinition_Versioning(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	for i := uint64(1); i <= 3; i++ {
		h := sha256.Sum256([]byte{byte(i)})
		err := store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
			Key: 1000 + i, BpmnProcessId: "proc", Version: i,
			ContentHash: h[:], Content: []byte{byte(i)}, DeployedAt: time.Now(),
		})
		require.NoError(t, err)
	}

	latest, err := store.ProcessDefinitions().FindLatestByProcessId(ctx, "proc")
	require.NoError(t, err)
	assert.Equal(t, uint64(3), latest.Version)

	v2, err := store.ProcessDefinitions().FindByProcessIdAndVersion(ctx, "proc", 2)
	require.NoError(t, err)
	require.NotNil(t, v2)
	assert.Equal(t, uint64(1002), v2.Key)
}

// --- ProcessInstance & ElementInstance ---

func TestProcessInstance_CRUD(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	pi := &storage.ProcessInstance{
		Key: 2001, ProcessDefinitionKey: 1001, BpmnProcessId: "proc",
		State: storage.ProcessInstanceActive, CreatedAt: time.Now().Truncate(time.Millisecond),
	}
	require.NoError(t, store.ProcessInstances().CreateInstance(ctx, pi))

	found, err := store.ProcessInstances().GetInstance(ctx, 2001)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, storage.ProcessInstanceActive, found.State)

	require.NoError(t, store.ProcessInstances().UpdateInstanceState(ctx, 2001, storage.ProcessInstanceCompleted))
	found, _ = store.ProcessInstances().GetInstance(ctx, 2001)
	assert.Equal(t, storage.ProcessInstanceCompleted, found.State)
}

func TestElementInstance_CRUD(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	ei := &storage.ElementInstance{
		Key: 3001, ProcessInstanceKey: 2001, ProcessDefinitionKey: 1001,
		ElementId: "start_1", ElementType: "startEvent", FlowScopeKey: 2001,
		State: storage.ElementInstanceActivated, CreatedAt: time.Now().Truncate(time.Millisecond),
	}
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, ei))

	found, err := store.ProcessInstances().GetElementInstance(ctx, 3001)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "start_1", found.ElementId)

	require.NoError(t, store.ProcessInstances().UpdateElementInstanceState(ctx, 3001, storage.ElementInstanceCompleted))
	found, _ = store.ProcessInstances().GetElementInstance(ctx, 3001)
	assert.Equal(t, storage.ElementInstanceCompleted, found.State)

	all, err := store.ProcessInstances().FindElementInstancesByProcessInstance(ctx, 2001)
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

// --- Variables ---

func TestVariable_ScopeOperations(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	// Create variables in scope
	require.NoError(t, store.Variables().Create(ctx, &storage.Variable{
		Key: 4001, ProcessInstanceKey: 2001, ScopeKey: 2001, Name: "x", Value: []byte(`42`),
	}))
	require.NoError(t, store.Variables().Create(ctx, &storage.Variable{
		Key: 4002, ProcessInstanceKey: 2001, ScopeKey: 2001, Name: "y", Value: []byte(`"hello"`),
	}))

	// FindByScope
	vars, err := store.Variables().FindByScope(ctx, 2001)
	require.NoError(t, err)
	assert.Len(t, vars, 2)

	// FindByName
	v, err := store.Variables().FindByName(ctx, 2001, "x")
	require.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, []byte(`42`), v.Value)

	// Update
	require.NoError(t, store.Variables().Update(ctx, 4001, []byte(`99`)))
	v, _ = store.Variables().FindByName(ctx, 2001, "x")
	assert.Equal(t, []byte(`99`), v.Value)

	// FindByName — not found
	v, err = store.Variables().FindByName(ctx, 2001, "z")
	require.NoError(t, err)
	assert.Nil(t, v)
}

// --- Jobs ---

func TestJob_Lifecycle(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	job := &storage.Job{
		Key: 5001, ProcessInstanceKey: 2001, ElementInstanceKey: 3001,
		ProcessDefinitionKey: 1001, Type: "payment-service",
		State: storage.JobCreated, Retries: 3, Variables: []byte(`{}`),
		CreatedAt: time.Now().Truncate(time.Millisecond),
	}
	require.NoError(t, store.Jobs().Create(ctx, job))

	// FindActivatable
	jobs, err := store.Jobs().FindActivatable(ctx, "payment-service", 10)
	require.NoError(t, err)
	assert.Len(t, jobs, 1)

	// Activate
	deadline := time.Now().Add(5 * time.Minute).Truncate(time.Millisecond)
	require.NoError(t, store.Jobs().Activate(ctx, 5001, "worker-1", deadline))

	found, _ := store.Jobs().GetByKey(ctx, 5001)
	assert.Equal(t, storage.JobActivated, found.State)
	assert.Equal(t, "worker-1", found.Worker)

	// No longer activatable
	jobs, _ = store.Jobs().FindActivatable(ctx, "payment-service", 10)
	assert.Empty(t, jobs)

	// Complete
	require.NoError(t, store.Jobs().Complete(ctx, 5001, []byte(`{"result":true}`)))
	found, _ = store.Jobs().GetByKey(ctx, 5001)
	assert.Equal(t, storage.JobCompleted, found.State)
}

func TestJob_FailAndThrowError(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	require.NoError(t, store.Jobs().Create(ctx, &storage.Job{
		Key: 5002, ProcessInstanceKey: 2001, ElementInstanceKey: 3001,
		ProcessDefinitionKey: 1001, Type: "test", State: storage.JobCreated,
		Retries: 3, CreatedAt: time.Now(),
	}))

	require.NoError(t, store.Jobs().Fail(ctx, 5002, 2, "timeout"))
	found, _ := store.Jobs().GetByKey(ctx, 5002)
	assert.Equal(t, storage.JobFailed, found.State)
	assert.Equal(t, 2, found.Retries)
	assert.Equal(t, "timeout", found.ErrorMessage)

	require.NoError(t, store.Jobs().ThrowError(ctx, 5002, "PAYMENT_FAILED", "card declined"))
	found, _ = store.Jobs().GetByKey(ctx, 5002)
	assert.Equal(t, storage.JobErrorThrown, found.State)
	assert.Equal(t, "PAYMENT_FAILED", found.ErrorCode)
}

// --- Timers ---

func TestTimer_FindDue(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	now := time.Now()
	// Past timer
	require.NoError(t, store.Timers().Create(ctx, &storage.Timer{
		Key: 6001, ProcessInstanceKey: 2001, ElementInstanceKey: 3001,
		ProcessDefinitionKey: 1001, State: storage.TimerCreated,
		DueDate: now.Add(-time.Minute), CreatedAt: now,
	}))
	// Future timer
	require.NoError(t, store.Timers().Create(ctx, &storage.Timer{
		Key: 6002, ProcessInstanceKey: 2001, ElementInstanceKey: 3002,
		ProcessDefinitionKey: 1001, State: storage.TimerCreated,
		DueDate: now.Add(time.Hour), CreatedAt: now,
	}))

	due, err := store.Timers().FindDue(ctx, now, 10)
	require.NoError(t, err)
	assert.Len(t, due, 1)
	assert.Equal(t, uint64(6001), due[0].Key)

	require.NoError(t, store.Timers().Trigger(ctx, 6001))
	due, _ = store.Timers().FindDue(ctx, now, 10)
	assert.Empty(t, due)
}

// --- Message Subscriptions ---

func TestMessageSubscription_Correlation(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	sub := &storage.MessageSubscription{
		Key: 7001, ProcessInstanceKey: 2001, ElementInstanceKey: 3001,
		MessageName: "order-paid", CorrelationKey: "order-123",
		State: storage.MessageSubscriptionOpened, CreatedAt: time.Now(),
	}
	require.NoError(t, store.MessageSubscriptions().CreateSubscription(ctx, sub))

	// Find open
	subs, err := store.MessageSubscriptions().FindOpenSubscriptions(ctx, "order-paid", "order-123")
	require.NoError(t, err)
	assert.Len(t, subs, 1)

	// Correlate
	require.NoError(t, store.MessageSubscriptions().Correlate(ctx, 7001))
	subs, _ = store.MessageSubscriptions().FindOpenSubscriptions(ctx, "order-paid", "order-123")
	assert.Empty(t, subs)
}

func TestMessageBuffer_TTL(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	now := time.Now()
	// Active message
	require.NoError(t, store.MessageSubscriptions().BufferMessage(ctx, &storage.MessageBuffer{
		Key: 8001, MessageName: "msg", CorrelationKey: "k1",
		Variables: []byte(`{}`), ExpiresAt: now.Add(time.Hour), CreatedAt: now,
	}))
	// Expired message
	require.NoError(t, store.MessageSubscriptions().BufferMessage(ctx, &storage.MessageBuffer{
		Key: 8002, MessageName: "msg", CorrelationKey: "k1",
		Variables: []byte(`{}`), ExpiresAt: now.Add(-time.Minute), CreatedAt: now,
	}))

	msgs, err := store.MessageSubscriptions().FindBufferedMessages(ctx, "msg", "k1")
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, uint64(8001), msgs[0].Key)

	cleaned, err := store.MessageSubscriptions().CleanExpired(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cleaned)
}

// --- Incidents ---

func TestIncident_CreateAndResolve(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	inc := &storage.Incident{
		Key: 9001, ProcessInstanceKey: 2001, ElementInstanceKey: 3001,
		Type: storage.IncidentTypeJobNoRetries, State: storage.IncidentCreated,
		ErrorMessage: "no retries left", CreatedAt: time.Now(),
	}
	require.NoError(t, store.Incidents().Create(ctx, inc))

	unresolved, err := store.Incidents().FindUnresolved(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, unresolved, 1)

	require.NoError(t, store.Incidents().Resolve(ctx, 9001))
	found, _ := store.Incidents().GetByKey(ctx, 9001)
	assert.Equal(t, storage.IncidentResolved, found.State)
	assert.False(t, found.ResolvedAt.IsZero())

	unresolved, _ = store.Incidents().FindUnresolved(ctx, 10)
	assert.Empty(t, unresolved)
}

// --- Transactional Execute ---

func TestStore_Execute_CommitOnSuccess(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	intents, err := store.Execute(ctx, storage.HandlerFunc(func(ctx context.Context, s storage.Store) ([]intent.Intent, error) {
		return nil, s.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
			Key: 1001, BpmnProcessId: "proc", Version: 1,
			ContentHash: []byte("h"), Content: []byte("c"), DeployedAt: time.Now(),
		})
	}))
	require.NoError(t, err)
	assert.Nil(t, intents)

	// Data persisted
	found, _ := store.ProcessDefinitions().FindByKey(ctx, 1001)
	assert.NotNil(t, found)
}

func TestStore_Execute_RollbackOnError(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()

	_, err := store.Execute(ctx, storage.HandlerFunc(func(ctx context.Context, s storage.Store) ([]intent.Intent, error) {
		_ = s.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
			Key: 1001, BpmnProcessId: "proc", Version: 1,
			ContentHash: []byte("h"), Content: []byte("c"), DeployedAt: time.Now(),
		})
		return nil, assert.AnError
	}))
	require.Error(t, err)

	// Data NOT persisted
	found, _ := store.ProcessDefinitions().FindByKey(ctx, 1001)
	assert.Nil(t, found)
}
