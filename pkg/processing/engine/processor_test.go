package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Store ---

type mockStore struct {
	procDefs     *mockProcessDefinitionRepo
	procInsts    *mockProcessInstanceRepo
	vars         *mockVariableRepo
	jobs         *mockJobRepo
	timers       *mockTimerRepo
	msgSubs      *mockMessageSubscriptionRepo
	incidents    *mockIncidentRepo
	executeCount int
	mu           sync.Mutex
}

func newMockStore() *mockStore {
	return &mockStore{
		procDefs:  &mockProcessDefinitionRepo{defs: make(map[uint64]*storage.ProcessDefinition)},
		procInsts: &mockProcessInstanceRepo{},
		vars:      &mockVariableRepo{},
		jobs:      &mockJobRepo{},
		timers:    &mockTimerRepo{},
		msgSubs:   &mockMessageSubscriptionRepo{},
		incidents: &mockIncidentRepo{},
	}
}

func (m *mockStore) ProcessDefinitions() storage.ProcessDefinitionRepository { return m.procDefs }
func (m *mockStore) ProcessInstances() storage.ProcessInstanceRepository     { return m.procInsts }
func (m *mockStore) Variables() storage.VariableRepository                   { return m.vars }
func (m *mockStore) Jobs() storage.JobRepository                             { return m.jobs }
func (m *mockStore) Timers() storage.TimerRepository                         { return m.timers }
func (m *mockStore) MessageSubscriptions() storage.MessageSubscriptionRepository {
	return m.msgSubs
}
func (m *mockStore) Incidents() storage.IncidentRepository { return m.incidents }
func (m *mockStore) Forms() storage.FormRepository         { return nil }

func (m *mockStore) Execute(ctx context.Context, h storage.Handler) ([]intent.Intent, error) {
	m.mu.Lock()
	m.executeCount++
	m.mu.Unlock()
	return h.Handle(ctx, m)
}

func (m *mockStore) getExecuteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.executeCount
}

// --- Minimal repo mocks ---

type mockProcessDefinitionRepo struct {
	defs    map[uint64]*storage.ProcessDefinition
	byHash  *storage.ProcessDefinition
	created []*storage.ProcessDefinition
	lastVer uint64
	mu      sync.Mutex
}

func (r *mockProcessDefinitionRepo) Create(_ context.Context, def *storage.ProcessDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defs[def.Key] = def
	r.created = append(r.created, def)
	return nil
}
func (r *mockProcessDefinitionRepo) FindByKey(_ context.Context, key uint64) (*storage.ProcessDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.defs[key], nil
}
func (r *mockProcessDefinitionRepo) FindLatestByProcessId(context.Context, string) (*storage.ProcessDefinition, error) {
	return nil, nil
}
func (r *mockProcessDefinitionRepo) FindByProcessIdAndVersion(context.Context, string, uint64) (*storage.ProcessDefinition, error) {
	return nil, nil
}
func (r *mockProcessDefinitionRepo) FindByContentHash(context.Context, string, []byte) (*storage.ProcessDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byHash, nil
}
func (r *mockProcessDefinitionRepo) GetLastVersion(context.Context, string) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastVer, nil
}
func (r *mockProcessDefinitionRepo) Delete(context.Context, uint64) error { return nil }

type mockProcessInstanceRepo struct{}

func (r *mockProcessInstanceRepo) CreateInstance(context.Context, *storage.ProcessInstance) error {
	return nil
}
func (r *mockProcessInstanceRepo) GetInstance(context.Context, uint64) (*storage.ProcessInstance, error) {
	return nil, nil
}
func (r *mockProcessInstanceRepo) UpdateInstanceState(context.Context, uint64, storage.ProcessInstanceState) error {
	return nil
}
func (r *mockProcessInstanceRepo) CreateElementInstance(context.Context, *storage.ElementInstance) error {
	return nil
}
func (r *mockProcessInstanceRepo) GetElementInstance(context.Context, uint64) (*storage.ElementInstance, error) {
	return nil, nil
}
func (r *mockProcessInstanceRepo) UpdateElementInstanceState(context.Context, uint64, storage.ElementInstanceState) error {
	return nil
}
func (r *mockProcessInstanceRepo) FindElementInstancesByProcessInstance(context.Context, uint64) ([]*storage.ElementInstance, error) {
	return nil, nil
}

type mockVariableRepo struct{}

func (r *mockVariableRepo) Create(context.Context, *storage.Variable) error { return nil }
func (r *mockVariableRepo) Update(context.Context, uint64, []byte) error    { return nil }
func (r *mockVariableRepo) FindByScope(context.Context, uint64) ([]*storage.Variable, error) {
	return nil, nil
}
func (r *mockVariableRepo) FindByName(context.Context, uint64, string) (*storage.Variable, error) {
	return nil, nil
}
func (r *mockVariableRepo) CopyToScope(context.Context, uint64, uint64, []string) error { return nil }
func (r *mockVariableRepo) PropagateToParent(context.Context, uint64, uint64, []string) error {
	return nil
}

type mockJobRepo struct{}

func (r *mockJobRepo) Create(context.Context, *storage.Job) error                { return nil }
func (r *mockJobRepo) GetByKey(context.Context, uint64) (*storage.Job, error)    { return nil, nil }
func (r *mockJobRepo) Activate(context.Context, uint64, string, time.Time) error { return nil }
func (r *mockJobRepo) Complete(context.Context, uint64, []byte) error            { return nil }
func (r *mockJobRepo) Fail(context.Context, uint64, int, string) error           { return nil }
func (r *mockJobRepo) ThrowError(context.Context, uint64, string, string) error  { return nil }
func (r *mockJobRepo) FindActivatable(context.Context, string, int) ([]*storage.Job, error) {
	return nil, nil
}
func (r *mockJobRepo) UpdateRetries(context.Context, uint64, int) error        { return nil }
func (r *mockJobRepo) UpdateDeadline(context.Context, uint64, time.Time) error { return nil }
func (r *mockJobRepo) FindByProcessInstance(context.Context, uint64) ([]*storage.Job, error) {
	return nil, nil
}

type mockTimerRepo struct{}

func (r *mockTimerRepo) Create(context.Context, *storage.Timer) error             { return nil }
func (r *mockTimerRepo) GetByKey(context.Context, uint64) (*storage.Timer, error) { return nil, nil }
func (r *mockTimerRepo) Trigger(context.Context, uint64) error                    { return nil }
func (r *mockTimerRepo) Cancel(context.Context, uint64) error                     { return nil }
func (r *mockTimerRepo) FindDue(context.Context, time.Time, int) ([]*storage.Timer, error) {
	return nil, nil
}
func (r *mockTimerRepo) FindByProcessInstance(context.Context, uint64) ([]*storage.Timer, error) {
	return nil, nil
}

type mockMessageSubscriptionRepo struct{}

func (r *mockMessageSubscriptionRepo) CreateSubscription(context.Context, *storage.MessageSubscription) error {
	return nil
}
func (r *mockMessageSubscriptionRepo) Correlate(context.Context, uint64) error { return nil }
func (r *mockMessageSubscriptionRepo) CloseSubscription(context.Context, uint64) error {
	return nil
}
func (r *mockMessageSubscriptionRepo) FindOpenSubscriptions(context.Context, string, string) ([]*storage.MessageSubscription, error) {
	return nil, nil
}
func (r *mockMessageSubscriptionRepo) FindByProcessInstance(context.Context, uint64) ([]*storage.MessageSubscription, error) {
	return nil, nil
}
func (r *mockMessageSubscriptionRepo) BufferMessage(context.Context, *storage.MessageBuffer) error {
	return nil
}
func (r *mockMessageSubscriptionRepo) FindBufferedMessages(context.Context, string, string) ([]*storage.MessageBuffer, error) {
	return nil, nil
}
func (r *mockMessageSubscriptionRepo) CleanExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type mockIncidentRepo struct {
	created []*storage.Incident
	mu      sync.Mutex
}

func (r *mockIncidentRepo) Create(_ context.Context, inc *storage.Incident) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.created = append(r.created, inc)
	return nil
}
func (r *mockIncidentRepo) GetByKey(context.Context, uint64) (*storage.Incident, error) {
	return nil, nil
}
func (r *mockIncidentRepo) Resolve(context.Context, uint64) error { return nil }
func (r *mockIncidentRepo) FindByProcessInstance(context.Context, uint64) ([]*storage.Incident, error) {
	return nil, nil
}
func (r *mockIncidentRepo) FindUnresolved(context.Context, int) ([]*storage.Incident, error) {
	return nil, nil
}

// --- Tests ---

func TestProcessor_DeployProcess(t *testing.T) {
	store := newMockStore()
	registry := behavior.DefaultRegistry()

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header: intent.Header{
			Origin: intent.External,
		},
		BpmnProcessId: "test-process",
		Name:          "Test Process",
		Content:       []byte("<xml>test</xml>"),
		ContentHash:   []byte("hash123"),
	})

	// Wait for processing
	require.Eventually(t, func() bool {
		return store.getExecuteCount() >= 1
	}, time.Second, 10*time.Millisecond)

	store.procDefs.mu.Lock()
	defer store.procDefs.mu.Unlock()
	require.Len(t, store.procDefs.created, 1)
	assert.Equal(t, "test-process", store.procDefs.created[0].BpmnProcessId)
	assert.Equal(t, uint64(1), store.procDefs.created[0].Version)
}

func TestProcessor_DeployProcess_Idempotent(t *testing.T) {
	store := newMockStore()
	// Simulate existing definition with same hash
	store.procDefs.byHash = &storage.ProcessDefinition{
		Key:           100,
		BpmnProcessId: "test-process",
		Version:       1,
	}
	registry := behavior.DefaultRegistry()

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header: intent.Header{
			Origin: intent.External,
		},
		BpmnProcessId: "test-process",
		Content:       []byte("<xml>test</xml>"),
		ContentHash:   []byte("hash123"),
	})

	require.Eventually(t, func() bool {
		return store.getExecuteCount() >= 1
	}, time.Second, 10*time.Millisecond)

	// No new definition created — idempotent
	store.procDefs.mu.Lock()
	defer store.procDefs.mu.Unlock()
	assert.Empty(t, store.procDefs.created)
}

func TestProcessor_KeyGeneration(t *testing.T) {
	kg := NewKeyGenerator(3, 0)

	k1 := kg.Next()
	k2 := kg.Next()

	assert.Equal(t, uint8(3), PartitionOf(k1))
	assert.Equal(t, uint8(3), PartitionOf(k2))
	assert.NotEqual(t, k1, k2)
	assert.Less(t, k1, k2)
}

func TestProcessor_QueuePriority(t *testing.T) {
	q := NewQueue(16, nil)

	// Enqueue low priority first
	low := &intent.TriggerTimerIntent{
		Header: intent.Header{Key: 1},
	}
	q.TryEnqueue(low)

	// Then high priority
	high := &intent.DeployProcessIntent{
		Header: intent.Header{Key: 2},
	}
	q.TryEnqueue(high)

	done := make(chan struct{})

	// High should come out first
	first, ok := q.Dequeue(done)
	require.True(t, ok)
	assert.Equal(t, intent.DeployProcess, first.IntentType())

	second, ok := q.Dequeue(done)
	require.True(t, ok)
	assert.Equal(t, intent.TriggerTimer, second.IntentType())
}

func TestProcessor_FollowUpIntents(t *testing.T) {
	store := newMockStore()

	// Custom registry with a behavior that produces follow-up
	registry := behavior.NewRegistry()
	registry.Register(intent.DeployProcess, behavior.Typed(
		func(_ context.Context, _ storage.Store, i *intent.DeployProcessIntent) ([]intent.Intent, error) {
			return []intent.Intent{
				&intent.CreateProcessInstanceIntent{
					Header: intent.Header{
						Origin: intent.Internal,
					},
					BpmnProcessId: i.BpmnProcessId,
				},
			}, nil
		},
	))

	var processedTypes []intent.Type
	var mu sync.Mutex

	registry.Register(intent.CreateProcessInstance, behavior.Typed(
		func(_ context.Context, _ storage.Store, i *intent.CreateProcessInstanceIntent) ([]intent.Intent, error) {
			mu.Lock()
			processedTypes = append(processedTypes, i.IntentType())
			mu.Unlock()
			return nil, nil
		},
	))

	proc := NewProcessor(Config{
		PartitionId: 1,
		Store:       store,
		Registry:    registry,
		QueueSize:   16,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go proc.Run(ctx)

	proc.Submit(&intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: "test",
	})

	// Wait for both intents to be processed
	require.Eventually(t, func() bool {
		return store.getExecuteCount() >= 2
	}, time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, processedTypes, 1)
	assert.Equal(t, intent.CreateProcessInstance, processedTypes[0])
}
