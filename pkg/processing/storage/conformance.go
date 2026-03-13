// Package storage provides a conformance test suite for storage.Store implementations.
//
// Any storage implementation can validate compatibility by calling RunConformanceSuite:
//
//	func TestMyStore(t *testing.T) {
//	    store := mypackage.NewStore(...)
//	    storage.RunConformanceSuite(t, store)
//	}
package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunConformanceSuite runs all conformance tests against a Store implementation.
// The store must be initialized (schema migrated) before calling this.
func RunConformanceSuite(t *testing.T, store Store) {
	t.Run("ProcessDefinition", func(t *testing.T) {
		runProcessDefinitionTests(t, store)
	})
	t.Run("ProcessInstance", func(t *testing.T) {
		runProcessInstanceTests(t, store)
	})
	t.Run("ElementInstance", func(t *testing.T) {
		runElementInstanceTests(t, store)
	})
	t.Run("Variable", func(t *testing.T) {
		runVariableTests(t, store)
	})
	t.Run("Job", func(t *testing.T) {
		runJobTests(t, store)
	})
	t.Run("Timer", func(t *testing.T) {
		runTimerTests(t, store)
	})
	t.Run("MessageSubscription", func(t *testing.T) {
		runMessageSubscriptionTests(t, store)
	})
	t.Run("Incident", func(t *testing.T) {
		runIncidentTests(t, store)
	})
	t.Run("Transaction", func(t *testing.T) {
		runTransactionTests(t, store)
	})
}

func runProcessDefinitionTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.ProcessDefinitions()

	t.Run("CreateAndFindByKey", func(t *testing.T) {
		hash := sha256.Sum256([]byte("content-1"))
		def := &ProcessDefinition{
			Key: 100001, BpmnProcessId: "conf-proc-1", Name: "Conformance Process 1",
			Version: 1, ContentHash: hash[:], Content: []byte("content-1"),
			DeployedAt: time.Now().Truncate(time.Millisecond),
		}
		require.NoError(t, repo.Create(ctx, def))

		found, err := repo.FindByKey(ctx, 100001)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "conf-proc-1", found.BpmnProcessId)
		assert.Equal(t, "Conformance Process 1", found.Name)
		assert.Equal(t, uint64(1), found.Version)
		assert.Equal(t, []byte("content-1"), found.Content)
	})

	t.Run("FindByKey_NotFound", func(t *testing.T) {
		found, err := repo.FindByKey(ctx, 999999)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindLatestByProcessId", func(t *testing.T) {
		for i := uint64(2); i <= 3; i++ {
			h := sha256.Sum256([]byte{byte(i), 'c'})
			require.NoError(t, repo.Create(ctx, &ProcessDefinition{
				Key: 100000 + i, BpmnProcessId: "conf-proc-1", Version: i,
				ContentHash: h[:], Content: []byte{byte(i)},
				DeployedAt: time.Now(),
			}))
		}
		found, err := repo.FindLatestByProcessId(ctx, "conf-proc-1")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, uint64(3), found.Version)
	})

	t.Run("FindByProcessIdAndVersion", func(t *testing.T) {
		found, err := repo.FindByProcessIdAndVersion(ctx, "conf-proc-1", 2)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, uint64(100002), found.Key)
	})

	t.Run("FindByContentHash", func(t *testing.T) {
		hash := sha256.Sum256([]byte("content-1"))
		found, err := repo.FindByContentHash(ctx, "conf-proc-1", hash[:])
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, uint64(100001), found.Key)
	})

	t.Run("FindByContentHash_NotFound", func(t *testing.T) {
		found, err := repo.FindByContentHash(ctx, "conf-proc-1", []byte("nonexistent"))
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("GetLastVersion", func(t *testing.T) {
		ver, err := repo.GetLastVersion(ctx, "conf-proc-1")
		require.NoError(t, err)
		assert.Equal(t, uint64(3), ver)
	})

	t.Run("GetLastVersion_Unknown", func(t *testing.T) {
		ver, err := repo.GetLastVersion(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), ver)
	})

	t.Run("Delete_SkipsInFindLatest", func(t *testing.T) {
		// Create two versions of a deletable process
		for i := uint64(1); i <= 2; i++ {
			h := sha256.Sum256([]byte(fmt.Sprintf("del-test-%d", i)))
			require.NoError(t, repo.Create(ctx, &ProcessDefinition{
				Key: 100100 + i, BpmnProcessId: "del-proc", Version: i,
				ContentHash: h[:], Content: []byte(fmt.Sprintf("del-content-%d", i)),
				DeployedAt: time.Now(),
			}))
		}
		// Delete the latest version
		require.NoError(t, repo.Delete(ctx, 100102))

		// FindLatest should return version 1 (skip deleted v2)
		found, err := repo.FindLatestByProcessId(ctx, "del-proc")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, uint64(1), found.Version)

		// FindByKey still works (returns deleted definition)
		deleted, err := repo.FindByKey(ctx, 100102)
		require.NoError(t, err)
		require.NotNil(t, deleted)
	})

	t.Run("Delete_AllVersions_FindLatestReturnsNil", func(t *testing.T) {
		h := sha256.Sum256([]byte("del-all-test"))
		require.NoError(t, repo.Create(ctx, &ProcessDefinition{
			Key: 100200, BpmnProcessId: "del-all-proc", Version: 1,
			ContentHash: h[:], Content: []byte("del-all-content"),
			DeployedAt: time.Now(),
		}))
		require.NoError(t, repo.Delete(ctx, 100200))

		found, err := repo.FindLatestByProcessId(ctx, "del-all-proc")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func runProcessInstanceTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.ProcessInstances()

	t.Run("CreateAndGet", func(t *testing.T) {
		pi := &ProcessInstance{
			Key: 200001, ProcessDefinitionKey: 100001, BpmnProcessId: "conf-proc-1",
			State: ProcessInstanceActive, CreatedAt: time.Now().Truncate(time.Millisecond),
		}
		require.NoError(t, repo.CreateInstance(ctx, pi))

		found, err := repo.GetInstance(ctx, 200001)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, ProcessInstanceActive, found.State)
		assert.Equal(t, "conf-proc-1", found.BpmnProcessId)
	})

	t.Run("GetInstance_NotFound", func(t *testing.T) {
		found, err := repo.GetInstance(ctx, 999999)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("UpdateState", func(t *testing.T) {
		require.NoError(t, repo.UpdateInstanceState(ctx, 200001, ProcessInstanceCompleted))
		found, _ := repo.GetInstance(ctx, 200001)
		assert.Equal(t, ProcessInstanceCompleted, found.State)
	})
}

func runElementInstanceTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.ProcessInstances()

	t.Run("CreateAndGet", func(t *testing.T) {
		ei := &ElementInstance{
			Key: 300001, ProcessInstanceKey: 200001, ProcessDefinitionKey: 100001,
			ElementId: "start_1", ElementType: "startEvent", FlowScopeKey: 200001,
			State: ElementInstanceActivated, CreatedAt: time.Now().Truncate(time.Millisecond),
		}
		require.NoError(t, repo.CreateElementInstance(ctx, ei))

		found, err := repo.GetElementInstance(ctx, 300001)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "start_1", found.ElementId)
		assert.Equal(t, ElementInstanceActivated, found.State)
	})

	t.Run("UpdateState", func(t *testing.T) {
		require.NoError(t, repo.UpdateElementInstanceState(ctx, 300001, ElementInstanceCompleted))
		found, _ := repo.GetElementInstance(ctx, 300001)
		assert.Equal(t, ElementInstanceCompleted, found.State)
	})

	t.Run("FindByProcessInstance", func(t *testing.T) {
		// Add another element
		require.NoError(t, repo.CreateElementInstance(ctx, &ElementInstance{
			Key: 300002, ProcessInstanceKey: 200001, ProcessDefinitionKey: 100001,
			ElementId: "task_1", ElementType: "serviceTask", FlowScopeKey: 200001,
			State: ElementInstanceActivated, CreatedAt: time.Now(),
		}))
		all, err := repo.FindElementInstancesByProcessInstance(ctx, 200001)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)
	})
}

func runVariableTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.Variables()

	t.Run("CreateAndFindByName", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, &Variable{
			Key: 400001, ProcessInstanceKey: 200001, ScopeKey: 200001,
			Name: "amount", Value: []byte(`100`),
		}))
		require.NoError(t, repo.Create(ctx, &Variable{
			Key: 400002, ProcessInstanceKey: 200001, ScopeKey: 200001,
			Name: "currency", Value: []byte(`"USD"`),
		}))

		v, err := repo.FindByName(ctx, 200001, "amount")
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.Equal(t, []byte(`100`), v.Value)
	})

	t.Run("FindByName_NotFound", func(t *testing.T) {
		v, err := repo.FindByName(ctx, 200001, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("FindByScope", func(t *testing.T) {
		vars, err := repo.FindByScope(ctx, 200001)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(vars), 2)
	})

	t.Run("Update", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, 400001, []byte(`200`)))
		v, _ := repo.FindByName(ctx, 200001, "amount")
		assert.Equal(t, []byte(`200`), v.Value)
	})
}

func runJobTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.Jobs()

	t.Run("CreateAndFind", func(t *testing.T) {
		job := &Job{
			Key: 500001, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			ProcessDefinitionKey: 100001, Type: "conf-worker",
			State: JobCreated, Retries: 3, Variables: []byte(`{}`),
			CreatedAt: time.Now().Truncate(time.Millisecond),
		}
		require.NoError(t, repo.Create(ctx, job))

		found, err := repo.GetByKey(ctx, 500001)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "conf-worker", found.Type)
		assert.Equal(t, JobCreated, found.State)
	})

	t.Run("FindActivatable", func(t *testing.T) {
		jobs, err := repo.FindActivatable(ctx, "conf-worker", 10)
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
	})

	t.Run("Activate", func(t *testing.T) {
		deadline := time.Now().Add(5 * time.Minute)
		require.NoError(t, repo.Activate(ctx, 500001, "worker-1", deadline))

		found, _ := repo.GetByKey(ctx, 500001)
		assert.Equal(t, JobActivated, found.State)
		assert.Equal(t, "worker-1", found.Worker)

		// No longer activatable
		jobs, _ := repo.FindActivatable(ctx, "conf-worker", 10)
		assert.Empty(t, jobs)
	})

	t.Run("UpdateDeadline", func(t *testing.T) {
		newDeadline := time.Now().Add(10 * time.Minute).Truncate(time.Millisecond)
		require.NoError(t, repo.UpdateDeadline(ctx, 500001, newDeadline))

		found, _ := repo.GetByKey(ctx, 500001)
		assert.Equal(t, newDeadline, found.Deadline)
	})

	t.Run("Complete", func(t *testing.T) {
		require.NoError(t, repo.Complete(ctx, 500001, []byte(`{"result":true}`)))
		found, _ := repo.GetByKey(ctx, 500001)
		assert.Equal(t, JobCompleted, found.State)
	})

	t.Run("Fail", func(t *testing.T) {
		// Create another job for fail test
		require.NoError(t, repo.Create(ctx, &Job{
			Key: 500002, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			ProcessDefinitionKey: 100001, Type: "conf-worker-2",
			State: JobCreated, Retries: 3, CreatedAt: time.Now(),
		}))
		require.NoError(t, repo.Fail(ctx, 500002, 2, "timeout"))
		found, _ := repo.GetByKey(ctx, 500002)
		assert.Equal(t, JobFailed, found.State)
		assert.Equal(t, 2, found.Retries)
		assert.Equal(t, "timeout", found.ErrorMessage)
	})

	t.Run("ThrowError", func(t *testing.T) {
		require.NoError(t, repo.ThrowError(ctx, 500002, "BIZ_ERR", "business error"))
		found, _ := repo.GetByKey(ctx, 500002)
		assert.Equal(t, JobErrorThrown, found.State)
		assert.Equal(t, "BIZ_ERR", found.ErrorCode)
	})

	t.Run("UpdateRetries", func(t *testing.T) {
		// Create a job to test retries update
		require.NoError(t, repo.Create(ctx, &Job{
			Key: 500010, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			ProcessDefinitionKey: 100001, Type: "conf-retries",
			State: JobCreated, Retries: 3, CreatedAt: time.Now(),
		}))
		require.NoError(t, repo.UpdateRetries(ctx, 500010, 5))
		found, _ := repo.GetByKey(ctx, 500010)
		assert.Equal(t, 5, found.Retries)
		assert.Equal(t, JobCreated, found.State) // state unchanged for non-failed jobs
	})

	t.Run("UpdateRetries_ReactivatesFailedJob", func(t *testing.T) {
		// Create and fail a job
		require.NoError(t, repo.Create(ctx, &Job{
			Key: 500011, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			ProcessDefinitionKey: 100001, Type: "conf-retries-fail",
			State: JobCreated, Retries: 1, CreatedAt: time.Now(),
		}))
		require.NoError(t, repo.Fail(ctx, 500011, 0, "exhausted"))
		found, _ := repo.GetByKey(ctx, 500011)
		assert.Equal(t, JobFailed, found.State)

		// UpdateRetries with > 0 should move back to Created
		require.NoError(t, repo.UpdateRetries(ctx, 500011, 3))
		found, _ = repo.GetByKey(ctx, 500011)
		assert.Equal(t, 3, found.Retries)
		assert.Equal(t, JobCreated, found.State)

		// Job should now be activatable
		jobs, err := repo.FindActivatable(ctx, "conf-retries-fail", 10)
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
	})

	t.Run("FindByProcessInstance", func(t *testing.T) {
		jobs, err := repo.FindByProcessInstance(ctx, 200001)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 2)
	})
}

func runTimerTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.Timers()
	now := time.Now()

	t.Run("CreateAndFindDue", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, &Timer{
			Key: 600001, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			ProcessDefinitionKey: 100001, State: TimerCreated,
			DueDate: now.Add(-time.Minute), CreatedAt: now,
		}))
		require.NoError(t, repo.Create(ctx, &Timer{
			Key: 600002, ProcessInstanceKey: 200001, ElementInstanceKey: 300002,
			ProcessDefinitionKey: 100001, State: TimerCreated,
			DueDate: now.Add(time.Hour), CreatedAt: now,
		}))

		due, err := repo.FindDue(ctx, now, 10)
		require.NoError(t, err)
		assert.Len(t, due, 1)
		assert.Equal(t, uint64(600001), due[0].Key)
	})

	t.Run("Trigger", func(t *testing.T) {
		require.NoError(t, repo.Trigger(ctx, 600001))
		due, _ := repo.FindDue(ctx, now, 10)
		assert.Empty(t, due)
	})

	t.Run("Cancel", func(t *testing.T) {
		require.NoError(t, repo.Cancel(ctx, 600002))
		found, _ := repo.GetByKey(ctx, 600002)
		assert.Equal(t, TimerCanceled, found.State)
	})
}

func runMessageSubscriptionTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.MessageSubscriptions()

	t.Run("CreateAndFind", func(t *testing.T) {
		sub := &MessageSubscription{
			Key: 700001, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			MessageName: "conf-msg", CorrelationKey: "order-42",
			State: MessageSubscriptionOpened, CreatedAt: time.Now(),
		}
		require.NoError(t, repo.CreateSubscription(ctx, sub))

		subs, err := repo.FindOpenSubscriptions(ctx, "conf-msg", "order-42")
		require.NoError(t, err)
		assert.Len(t, subs, 1)
	})

	t.Run("Correlate", func(t *testing.T) {
		require.NoError(t, repo.Correlate(ctx, 700001))
		subs, _ := repo.FindOpenSubscriptions(ctx, "conf-msg", "order-42")
		assert.Empty(t, subs)
	})

	t.Run("BufferAndClean", func(t *testing.T) {
		now := time.Now()
		require.NoError(t, repo.BufferMessage(ctx, &MessageBuffer{
			Key: 800001, MessageName: "conf-msg", CorrelationKey: "order-42",
			Variables: []byte(`{}`), ExpiresAt: now.Add(time.Hour), CreatedAt: now,
		}))
		require.NoError(t, repo.BufferMessage(ctx, &MessageBuffer{
			Key: 800002, MessageName: "conf-msg", CorrelationKey: "order-42",
			Variables: []byte(`{}`), ExpiresAt: now.Add(-time.Minute), CreatedAt: now,
		}))

		msgs, err := repo.FindBufferedMessages(ctx, "conf-msg", "order-42")
		require.NoError(t, err)
		assert.Len(t, msgs, 1)

		cleaned, err := repo.CleanExpired(ctx, now)
		require.NoError(t, err)
		assert.Equal(t, int64(1), cleaned)
	})
}

func runIncidentTests(t *testing.T, store Store) {
	ctx := context.Background()
	repo := store.Incidents()

	t.Run("CreateAndFind", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, &Incident{
			Key: 900001, ProcessInstanceKey: 200001, ElementInstanceKey: 300001,
			Type: IncidentTypeJobNoRetries, State: IncidentCreated,
			ErrorMessage: "no retries", CreatedAt: time.Now(),
		}))

		unresolved, err := repo.FindUnresolved(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(unresolved), 1)
	})

	t.Run("Resolve", func(t *testing.T) {
		require.NoError(t, repo.Resolve(ctx, 900001))
		found, _ := repo.GetByKey(ctx, 900001)
		assert.Equal(t, IncidentResolved, found.State)
		assert.False(t, found.ResolvedAt.IsZero())
	})

	t.Run("FindByProcessInstance", func(t *testing.T) {
		incs, err := repo.FindByProcessInstance(ctx, 200001)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(incs), 1)
	})
}

func runTransactionTests(t *testing.T, store Store) {
	ctx := context.Background()

	t.Run("CommitOnSuccess", func(t *testing.T) {
		hash := sha256.Sum256([]byte("tx-commit-test"))
		_, err := store.Execute(ctx, HandlerFunc(func(ctx context.Context, s Store) ([]intent.Intent, error) {
			return nil, s.ProcessDefinitions().Create(ctx, &ProcessDefinition{
				Key: 990001, BpmnProcessId: "tx-test", Version: 1,
				ContentHash: hash[:], Content: []byte("tx-commit-test"),
				DeployedAt: time.Now(),
			})
		}))
		require.NoError(t, err)

		found, err := store.ProcessDefinitions().FindByKey(ctx, 990001)
		require.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("RollbackOnError", func(t *testing.T) {
		hash := sha256.Sum256([]byte("tx-rollback-test"))
		_, err := store.Execute(ctx, HandlerFunc(func(ctx context.Context, s Store) ([]intent.Intent, error) {
			_ = s.ProcessDefinitions().Create(ctx, &ProcessDefinition{
				Key: 990002, BpmnProcessId: "tx-rollback", Version: 1,
				ContentHash: hash[:], Content: []byte("tx-rollback-test"),
				DeployedAt: time.Now(),
			})
			return nil, fmt.Errorf("intentional error")
		}))
		require.Error(t, err)

		found, err := store.ProcessDefinitions().FindByKey(ctx, 990002)
		require.NoError(t, err)
		assert.Nil(t, found, "data should NOT be persisted after rollback")
	})
}
