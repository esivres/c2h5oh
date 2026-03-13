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

func seedProcessInstance(t *testing.T, s storage.Store, piKey uint64, state storage.ProcessInstanceState) {
	t.Helper()
	require.NoError(t, s.ProcessInstances().CreateInstance(context.Background(), &storage.ProcessInstance{
		Key:                  piKey,
		ProcessDefinitionKey: 1,
		BpmnProcessId:        "test-process",
		State:                state,
		CreatedAt:            time.Now(),
	}))
}

func seedElementInstance(t *testing.T, s storage.Store, key, piKey uint64, elemType string, state storage.ElementInstanceState) {
	t.Helper()
	require.NoError(t, s.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                key,
		ProcessInstanceKey: piKey,
		ElementId:          "elem-" + elemType,
		ElementType:        elemType,
		FlowScopeKey:       piKey,
		State:              state,
		CreatedAt:          time.Now(),
	}))
}

// --- CancelProcessInstance ---

func TestCancelProcessInstance_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedProcessInstance(t, store, 100, storage.ProcessInstanceActive)
	seedElementInstance(t, store, 201, 100, "serviceTask", storage.ElementInstanceActivated)
	seedElementInstance(t, store, 202, 100, "startEvent", storage.ElementInstanceCompleted)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(cancelProcessInstance),
		&intent.CancelProcessInstanceIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
		},
	))
	require.NoError(t, err)

	// Should only terminate active elements (201), not completed ones (202)
	require.Len(t, intents, 1)
	term, ok := intents[0].(*intent.TerminateElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(201), term.ElementInstanceKey)

	// Process instance should be terminated
	pi, err := store.ProcessInstances().GetInstance(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, storage.ProcessInstanceTerminated, pi.State)
}

func TestCancelProcessInstance_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(cancelProcessInstance),
		&intent.CancelProcessInstanceIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 999},
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "process instance not found")
}

func TestCancelProcessInstance_AlreadyCompleted(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedProcessInstance(t, store, 100, storage.ProcessInstanceCompleted)

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(cancelProcessInstance),
		&intent.CancelProcessInstanceIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

// --- TerminateElement ---

func TestTerminateElement_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedProcessInstance(t, store, 100, storage.ProcessInstanceActive)
	seedElementInstance(t, store, 201, 100, "serviceTask", storage.ElementInstanceActivated)

	// Create a job for the element
	require.NoError(t, store.Jobs().Create(context.Background(), &storage.Job{
		Key:                1,
		ProcessInstanceKey: 100,
		ElementInstanceKey: 201,
		Type:               "worker",
		State:              storage.JobCreated,
		Retries:            3,
		CreatedAt:          time.Now(),
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(terminateElement),
		&intent.TerminateElementIntent{
			Header:             intent.Header{Key: 10, ProcessInstanceKey: 100},
			ElementInstanceKey: 201,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	// Element should be terminated
	ei, err := store.ProcessInstances().GetElementInstance(context.Background(), 201)
	require.NoError(t, err)
	assert.Equal(t, storage.ElementInstanceTerminated, ei.State)

	// Job should be failed with 0 retries
	job, err := store.Jobs().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.JobFailed, job.State)
	assert.Equal(t, 0, job.Retries)
}

func TestTerminateElement_NonServiceTask(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedProcessInstance(t, store, 100, storage.ProcessInstanceActive)
	seedElementInstance(t, store, 201, 100, "exclusiveGateway", storage.ElementInstanceActivated)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(terminateElement),
		&intent.TerminateElementIntent{
			Header:             intent.Header{Key: 10, ProcessInstanceKey: 100},
			ElementInstanceKey: 201,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	ei, err := store.ProcessInstances().GetElementInstance(context.Background(), 201)
	require.NoError(t, err)
	assert.Equal(t, storage.ElementInstanceTerminated, ei.State)
}

func TestTerminateElement_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(terminateElement),
		&intent.TerminateElementIntent{
			Header:             intent.Header{Key: 10, ProcessInstanceKey: 100},
			ElementInstanceKey: 999,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "element instance not found")
}
