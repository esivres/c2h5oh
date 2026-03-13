package behavior

import (
	"context"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"database/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, sqlstore.Migrate(context.Background(), db))
	return db
}

func seedJob(t *testing.T, s storage.Store, key uint64, state storage.JobState, retries int) *storage.Job {
	t.Helper()
	return seedJobFull(t, s, key, 100, 200, 0, state, retries)
}

func seedJobFull(t *testing.T, s storage.Store, key, piKey, eiKey, pdKey uint64, state storage.JobState, retries int) *storage.Job {
	t.Helper()
	job := &storage.Job{
		Key:                  key,
		ProcessInstanceKey:   piKey,
		ElementInstanceKey:   eiKey,
		ProcessDefinitionKey: pdKey,
		Type:                 "test-worker",
		State:                state,
		Retries:              retries,
		CreatedAt:            time.Now(),
	}
	require.NoError(t, s.Jobs().Create(context.Background(), job))
	if state == storage.JobActivated {
		require.NoError(t, s.Jobs().Activate(context.Background(), key, "worker-1", time.Now().Add(time.Minute)))
	}
	return job
}

func seedElementInstanceFull(t *testing.T, s storage.Store, key, piKey, pdKey, flowScopeKey uint64, elementId, elementType string) {
	t.Helper()
	require.NoError(t, s.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  key,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            elementId,
		ElementType:          elementType,
		FlowScopeKey:         flowScopeKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))
}

// --- ActivateJob ---

func TestActivateJob_Success(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedJob(t, store, 1, storage.JobCreated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(activateJob),
		&intent.ActivateJobIntent{
			Header:  intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey:  1,
			Worker:  "my-worker",
			Timeout: 30 * time.Second,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents)

	job, err := store.Jobs().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.JobActivated, job.State)
	assert.Equal(t, "my-worker", job.Worker)
	assert.False(t, job.Deadline.IsZero())
}

func TestActivateJob_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(activateJob),
		&intent.ActivateJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 999,
			Worker: "worker",
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestActivateJob_AlreadyActivated(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedJob(t, store, 1, storage.JobActivated, 3)

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(activateJob),
		&intent.ActivateJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 1,
			Worker: "worker",
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not activatable")
}

// --- ThrowJobError ---

func TestThrowJobError_CreatesIncident(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedElementInstanceFull(t, store, 200, 100, 0, 0, "task1", "serviceTask")
	seedJob(t, store, 1, storage.JobActivated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(throwJobError),
		&intent.ThrowJobErrorIntent{
			Header:       intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey:       1,
			ErrorCode:    "PAYMENT_FAILED",
			ErrorMessage: "insufficient funds",
		},
	))
	require.NoError(t, err)

	// No boundary event → create incident
	require.Len(t, intents, 1)
	inc, ok := intents[0].(*intent.CreateIncidentIntent)
	require.True(t, ok)
	assert.Equal(t, "UNHANDLED_BPMN_ERROR", inc.ErrorType)
	assert.Contains(t, inc.ErrorMessage, "PAYMENT_FAILED")

	// Job should be in ErrorThrown state
	job, err := store.Jobs().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.JobErrorThrown, job.State)
	assert.Equal(t, "PAYMENT_FAILED", job.ErrorCode)
}

func TestThrowJobError_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(throwJobError),
		&intent.ThrowJobErrorIntent{
			Header:    intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey:    999,
			ErrorCode: "ERR",
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

// --- TimeoutJob ---

func TestTimeoutJob_WithRetries(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedJob(t, store, 1, storage.JobActivated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(timeoutJob),
		&intent.TimeOutJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 1,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents) // retries decremented, job returned to Failed state for re-activation

	job, err := store.Jobs().GetByKey(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, storage.JobFailed, job.State)
	assert.Equal(t, 2, job.Retries) // 3 - 1 = 2
}

func TestTimeoutJob_NoRetriesLeft(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedJob(t, store, 1, storage.JobActivated, 1)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(timeoutJob),
		&intent.TimeOutJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 1,
		},
	))
	require.NoError(t, err)

	// Should emit FailJobIntent with retries=0
	require.Len(t, intents, 1)
	fail, ok := intents[0].(*intent.FailJobIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(1), fail.JobKey)
	assert.Equal(t, 0, fail.Retries)
}

func TestTimeoutJob_NotActivated(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedJob(t, store, 1, storage.JobCreated, 3)

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(timeoutJob),
		&intent.TimeOutJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 1,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not activated")
}

func TestTimeoutJob_NotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	_, err := store.Execute(context.Background(), AsHandler(
		Typed(timeoutJob),
		&intent.TimeOutJobIntent{
			Header: intent.Header{Key: 10, ProcessInstanceKey: 100},
			JobKey: 999,
		},
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

// --- ThrowJobError with Boundary Error Events ---

const bpmnWithBoundaryError = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:error id="Error_1" name="PaymentError" errorCode="PAYMENT_FAILED"/>
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="start"/>
    <bpmn:serviceTask id="task1" name="Process Payment"/>
    <bpmn:boundaryEvent id="boundary1" attachedToRef="task1" cancelActivity="true">
      <bpmn:errorEventDefinition id="errDef1" errorRef="Error_1"/>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="end1"/>
    <bpmn:endEvent id="errorEnd"/>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="end1"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="boundary1" targetRef="errorEnd"/>
  </bpmn:process>
</bpmn:definitions>`

const bpmnWithCatchAllBoundaryError = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:serviceTask id="task1" name="Do Work"/>
    <bpmn:boundaryEvent id="boundary1" attachedToRef="task1" cancelActivity="true">
      <bpmn:errorEventDefinition id="errDef1"/>
    </bpmn:boundaryEvent>
  </bpmn:process>
</bpmn:definitions>`

func seedBpmnDefinition(t *testing.T, s storage.Store, key uint64, bpmnXml string) {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnXml)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, s.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           key,
		BpmnProcessId: "Process_1",
		Version:       1,
		Content:       content,
		ContentHash:   []byte("test-hash"),
		DeployedAt:    time.Now(),
	}))
}

func TestThrowJobError_BoundaryEventMatched(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	var pdKey uint64 = 10
	seedBpmnDefinition(t, store, pdKey, bpmnWithBoundaryError)
	seedElementInstanceFull(t, store, 200, 100, pdKey, 100, "task1", "serviceTask")
	seedJobFull(t, store, 1, 100, 200, pdKey, storage.JobActivated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(throwJobError),
		&intent.ThrowJobErrorIntent{
			Header:       intent.Header{Key: 11, ProcessInstanceKey: 100},
			JobKey:       1,
			ErrorCode:    "PAYMENT_FAILED",
			ErrorMessage: "insufficient funds",
		},
	))
	require.NoError(t, err)

	// Should find boundary error event → TerminateElement + ActivateElement
	require.Len(t, intents, 2)

	term, ok := intents[0].(*intent.TerminateElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(200), term.ElementInstanceKey)

	activate, ok := intents[1].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "boundary1", activate.ElementId)
	assert.Equal(t, "boundaryEvent", activate.ElementType)
	assert.Equal(t, uint64(100), activate.FlowScopeKey)
}

func TestThrowJobError_CatchAllBoundaryEvent(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	var pdKey uint64 = 10
	seedBpmnDefinition(t, store, pdKey, bpmnWithCatchAllBoundaryError)
	seedElementInstanceFull(t, store, 200, 100, pdKey, 100, "task1", "serviceTask")
	seedJobFull(t, store, 1, 100, 200, pdKey, storage.JobActivated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(throwJobError),
		&intent.ThrowJobErrorIntent{
			Header:       intent.Header{Key: 11, ProcessInstanceKey: 100},
			JobKey:       1,
			ErrorCode:    "ANY_ERROR",
			ErrorMessage: "some error",
		},
	))
	require.NoError(t, err)

	// Catch-all boundary event (no errorRef) should match any error
	require.Len(t, intents, 2)
	activate, ok := intents[1].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "boundary1", activate.ElementId)
}

func TestThrowJobError_NoMatchingBoundaryEvent(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	var pdKey uint64 = 10
	seedBpmnDefinition(t, store, pdKey, bpmnWithBoundaryError)
	seedElementInstanceFull(t, store, 200, 100, pdKey, 100, "task1", "serviceTask")
	seedJobFull(t, store, 1, 100, 200, pdKey, storage.JobActivated, 3)

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(throwJobError),
		&intent.ThrowJobErrorIntent{
			Header:       intent.Header{Key: 11, ProcessInstanceKey: 100},
			JobKey:       1,
			ErrorCode:    "UNKNOWN_ERROR",
			ErrorMessage: "no matching boundary",
		},
	))
	require.NoError(t, err)

	// Error code doesn't match → incident
	require.Len(t, intents, 1)
	inc, ok := intents[0].(*intent.CreateIncidentIntent)
	require.True(t, ok)
	assert.Equal(t, "UNHANDLED_BPMN_ERROR", inc.ErrorType)
}
