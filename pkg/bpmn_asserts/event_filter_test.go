package bpmn_asserts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Generic filters ---

func TestEventFilter_ValueType(t *testing.T) {
	pred := Events().ValueType(ValueTypeJob).RecordPredicate()

	assert.True(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)))
	assert.False(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT")))
}

func TestEventFilter_Intent(t *testing.T) {
	pred := Events().Intent(IntentCompleted).RecordPredicate()

	assert.True(t, pred(makeJobRecord("COMPLETED", 1, "myJob", "task1", 100, 3)))
	assert.False(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)))
}

func TestEventFilter_RecordType(t *testing.T) {
	pred := Events().RecordType(RecordTypeEvent).RecordPredicate()

	assert.True(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)))

	cmdRecord := makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)
	cmdRecord.RecordType = RecordTypeCommand
	assert.False(t, pred(cmdRecord))
}

func TestEventFilter_WithKey(t *testing.T) {
	pred := Events().WithKey(42).RecordPredicate()

	rec := makeJobRecord("CREATED", 42, "myJob", "task1", 100, 3)
	assert.True(t, pred(rec))

	rec2 := makeJobRecord("CREATED", 99, "myJob", "task1", 100, 3)
	assert.False(t, pred(rec2))
}

func TestEventFilter_Where(t *testing.T) {
	pred := Events().Where(func(r Record) bool {
		return r.Key > 10
	}).RecordPredicate()

	assert.True(t, pred(Record{Key: 11}))
	assert.False(t, pred(Record{Key: 5}))
}

func TestEventFilter_CombinedFilters(t *testing.T) {
	pred := Events().
		ValueType("JOB").
		RecordType("EVENT").
		Intent("CREATED").
		RecordPredicate()

	assert.True(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)))
	assert.False(t, pred(makeJobRecord("COMPLETED", 1, "myJob", "task1", 100, 3)))
	assert.False(t, pred(makeProcessInstanceRecord("CREATED", 1, 100, "start", "START_EVENT")))
}

// --- Value-specific filters ---

func TestEventFilter_ProcessInstanceKey(t *testing.T) {
	pred := Events().ProcessInstanceKey(100).RecordPredicate()

	assert.True(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT")))
	assert.False(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 200, "start", "START_EVENT")))
}

func TestEventFilter_ProcessInstanceKey_OnJob(t *testing.T) {
	pred := Events().ProcessInstanceKey(100).RecordPredicate()

	assert.True(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)))
	assert.False(t, pred(makeJobRecord("CREATED", 1, "myJob", "task1", 200, 3)))
}

func TestEventFilter_ElementID(t *testing.T) {
	pred := Events().ElementID("task1").RecordPredicate()

	assert.True(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK")))
	assert.False(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task2", "SERVICE_TASK")))
}

func TestEventFilter_ElementType(t *testing.T) {
	pred := Events().ElementType("PROCESS").RecordPredicate()

	assert.True(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "proc", "PROCESS")))
	assert.False(t, pred(makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK")))
}

func TestEventFilter_BpmnProcessID(t *testing.T) {
	// BpmnProcessID filter uses DeploymentValue with ProcessesMetadata
	pred := Events().BpmnProcessID("myProcess").RecordPredicate()

	depRec := makeDeploymentRecord("CREATED", 1, []string{"myProcess"}, []string{"proc.bpmn"})
	assert.True(t, pred(depRec))

	depRec2 := makeDeploymentRecord("CREATED", 2, []string{"otherProcess"}, []string{"other.bpmn"})
	assert.False(t, pred(depRec2))
}

func TestEventFilter_VariableName(t *testing.T) {
	pred := Events().VariableName("myVar").RecordPredicate()

	assert.True(t, pred(makeVariableRecord("CREATED", 1, 100, "myVar", `"hello"`)))
	assert.False(t, pred(makeVariableRecord("CREATED", 1, 100, "otherVar", `"hello"`)))
}

func TestEventFilter_JobType(t *testing.T) {
	pred := Events().JobType("kamundarf:adhoc:v1").RecordPredicate()

	assert.True(t, pred(makeJobRecord("CREATED", 1, "kamundarf:adhoc:v1", "task1", 100, 3)))
	assert.False(t, pred(makeJobRecord("CREATED", 1, "someOtherType", "task1", 100, 3)))
}

func TestEventFilter_MessageName(t *testing.T) {
	pred := Events().MessageName("paymentReceived").RecordPredicate()

	subRec := makeMessageSubscriptionRecord("CORRELATED", 1, 100, 50, "paymentReceived")
	assert.True(t, pred(subRec))

	subRec2 := makeMessageSubscriptionRecord("CORRELATED", 1, 100, 50, "otherMessage")
	assert.False(t, pred(subRec2))
}

func TestEventFilter_MessageName_OnMessageRecord(t *testing.T) {
	pred := Events().MessageName("paymentReceived").RecordPredicate()

	// MessageValue uses "name" field, not "messageName"
	msgRec := makeMessageRecord("PUBLISHED", 1, "paymentReceived", "corr-123")
	assert.True(t, pred(msgRec))
}

func TestEventFilter_ErrorType(t *testing.T) {
	pred := Events().ErrorType("JOB_NO_RETRIES").RecordPredicate()

	assert.True(t, pred(makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "msg", "task1", 50)))
	assert.False(t, pred(makeIncidentRecord("CREATED", 1, 100, "CONDITION_ERROR", "msg", "task1", 50)))
}

// --- Shorthand builders ---

func TestEventFilter_ProcessInstanceCompleted(t *testing.T) {
	pred := Events().ProcessInstanceCompleted(100).RecordPredicate()

	// Should match: PROCESS_INSTANCE + EVENT + ELEMENT_COMPLETED + processInstanceKey=100 + elementType=PROCESS
	completedProcess := makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "testProcess", "PROCESS")
	assert.True(t, pred(completedProcess))

	// Should NOT match: element completed but not PROCESS type
	completedTask := makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task1", "SERVICE_TASK")
	assert.False(t, pred(completedTask))

	// Should NOT match: different process instance key
	otherProcess := makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 200, "testProcess", "PROCESS")
	assert.False(t, pred(otherProcess))

	// Should NOT match: activated, not completed
	activated := makeProcessInstanceRecord("ELEMENT_ACTIVATED", 4, 100, "testProcess", "PROCESS")
	assert.False(t, pred(activated))
}

func TestEventFilter_ProcessInstanceActivated(t *testing.T) {
	pred := Events().ProcessInstanceActivated(100).RecordPredicate()

	activated := makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS")
	assert.True(t, pred(activated))

	completedProcess := makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS")
	assert.False(t, pred(completedProcess))
}

func TestEventFilter_ProcessInstanceTerminated(t *testing.T) {
	pred := Events().ProcessInstanceTerminated(100).RecordPredicate()

	terminated := makeProcessInstanceRecord("ELEMENT_TERMINATED", 1, 100, "testProcess", "PROCESS")
	assert.True(t, pred(terminated))
}

func TestEventFilter_ElementCompleted(t *testing.T) {
	pred := Events().ElementCompleted(100, "task1").RecordPredicate()

	match := makeProcessInstanceRecord("ELEMENT_COMPLETED", 1, 100, "task1", "SERVICE_TASK")
	assert.True(t, pred(match))

	wrongElement := makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task2", "SERVICE_TASK")
	assert.False(t, pred(wrongElement))

	wrongKey := makeProcessInstanceRecord("ELEMENT_COMPLETED", 3, 200, "task1", "SERVICE_TASK")
	assert.False(t, pred(wrongKey))
}

func TestEventFilter_ElementActivated(t *testing.T) {
	pred := Events().ElementActivated(100, "task1").RecordPredicate()

	match := makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "task1", "SERVICE_TASK")
	assert.True(t, pred(match))

	wrongIntent := makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "task1", "SERVICE_TASK")
	assert.False(t, pred(wrongIntent))
}

func TestEventFilter_JobCreated(t *testing.T) {
	pred := Events().JobCreated("kamundarf:adhoc:v1").RecordPredicate()

	match := makeJobRecord("CREATED", 1, "kamundarf:adhoc:v1", "task1", 100, 3)
	assert.True(t, pred(match))

	wrongIntent := makeJobRecord("COMPLETED", 2, "kamundarf:adhoc:v1", "task1", 100, 3)
	assert.False(t, pred(wrongIntent))

	wrongType := makeJobRecord("CREATED", 3, "otherType", "task1", 100, 3)
	assert.False(t, pred(wrongType))
}

func TestEventFilter_JobCompleted(t *testing.T) {
	pred := Events().JobCompleted("kamundarf:adhoc:v1").RecordPredicate()

	match := makeJobRecord("COMPLETED", 1, "kamundarf:adhoc:v1", "task1", 100, 3)
	assert.True(t, pred(match))

	wrongIntent := makeJobRecord("CREATED", 2, "kamundarf:adhoc:v1", "task1", 100, 3)
	assert.False(t, pred(wrongIntent))
}

func TestEventFilter_IncidentCreated(t *testing.T) {
	pred := Events().IncidentCreated(100).RecordPredicate()

	match := makeIncidentRecord("CREATED", 1, 100, "JOB_NO_RETRIES", "msg", "task1", 50)
	assert.True(t, pred(match))

	wrongKey := makeIncidentRecord("CREATED", 2, 200, "JOB_NO_RETRIES", "msg", "task1", 50)
	assert.False(t, pred(wrongKey))

	wrongIntent := makeIncidentRecord("RESOLVED", 3, 100, "JOB_NO_RETRIES", "msg", "task1", 50)
	assert.False(t, pred(wrongIntent))
}

func TestEventFilter_IncidentResolved(t *testing.T) {
	pred := Events().IncidentResolved(100).RecordPredicate()

	match := makeIncidentRecord("RESOLVED", 1, 100, "JOB_NO_RETRIES", "msg", "task1", 50)
	assert.True(t, pred(match))
}

func TestEventFilter_VariableCreated(t *testing.T) {
	pred := Events().VariableCreated(100, "result").RecordPredicate()

	match := makeVariableRecord("CREATED", 1, 100, "result", `"ok"`)
	assert.True(t, pred(match))

	wrongName := makeVariableRecord("CREATED", 2, 100, "input", `"ok"`)
	assert.False(t, pred(wrongName))

	wrongKey := makeVariableRecord("CREATED", 3, 200, "result", `"ok"`)
	assert.False(t, pred(wrongKey))
}

func TestEventFilter_VariableUpdated(t *testing.T) {
	pred := Events().VariableUpdated(100, "result").RecordPredicate()

	match := makeVariableRecord("UPDATED", 1, 100, "result", `"new"`)
	assert.True(t, pred(match))

	wrongIntent := makeVariableRecord("CREATED", 2, 100, "result", `"ok"`)
	assert.False(t, pred(wrongIntent))
}

func TestEventFilter_MessageCorrelated(t *testing.T) {
	pred := Events().MessageCorrelated(100, "paymentReceived").RecordPredicate()

	match := makeMessageSubscriptionRecord("CORRELATED", 1, 100, 50, "paymentReceived")
	assert.True(t, pred(match))

	wrongMsg := makeMessageSubscriptionRecord("CORRELATED", 2, 100, 50, "otherMsg")
	assert.False(t, pred(wrongMsg))

	wrongKey := makeMessageSubscriptionRecord("CORRELATED", 3, 200, 50, "paymentReceived")
	assert.False(t, pred(wrongKey))
}

func TestEventFilter_DeploymentCreated(t *testing.T) {
	pred := Events().DeploymentCreated().RecordPredicate()

	match := makeDeploymentRecord("CREATED", 1, []string{"myProcess"}, []string{"proc.bpmn"})
	assert.True(t, pred(match))

	wrongIntent := makeDeploymentRecord("DISTRIBUTE", 2, []string{"myProcess"}, []string{"proc.bpmn"})
	assert.False(t, pred(wrongIntent))
}

// --- Output methods ---

func TestEventFilter_RawPredicate(t *testing.T) {
	rawPred := Events().ValueType(ValueTypeJob).Intent(IntentCreated).RawPredicate()

	rec := makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3)
	raw, _ := json.Marshal(rec)
	assert.True(t, rawPred(raw))

	rec2 := makeJobRecord("COMPLETED", 2, "myJob", "task1", 100, 3)
	raw2, _ := json.Marshal(rec2)
	assert.False(t, rawPred(raw2))
}

func TestEventFilter_RawPredicate_InvalidJSON(t *testing.T) {
	rawPred := Events().ValueType(ValueTypeJob).RawPredicate()
	assert.False(t, rawPred(json.RawMessage(`not json`)))
}

func TestEventFilter_WaitOn(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3),
	)

	rec, err := Events().JobCreated("myJob").WaitOn(stream, 100*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rec.Key)
}

func TestEventFilter_WaitOn_Timeout(t *testing.T) {
	stream := NewRecordStream()

	_, err := Events().JobCreated("myJob").WaitOn(stream, 50*time.Millisecond)
	assert.Error(t, err)
}

func TestEventFilter_WaitOnCtx(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	rec, err := Events().JobCreated("myJob").WaitOnCtx(stream, ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rec.Key)
}

func TestEventFilter_FindAll(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3),
		makeJobRecord("CREATED", 2, "myJob", "task2", 100, 3),
		makeJobRecord("COMPLETED", 3, "myJob", "task1", 100, 3),
	)

	records := Events().ValueType(ValueTypeJob).Intent(IntentCreated).FindAll(stream)
	assert.Len(t, records, 2)
}

func TestEventFilter_FindFirst(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3),
		makeJobRecord("CREATED", 2, "myJob", "task2", 100, 3),
	)

	rec := Events().ValueType(ValueTypeJob).FindFirst(stream)
	require.NotNil(t, rec)
	assert.Equal(t, int64(1), rec.Key)
}

func TestEventFilter_FindFirst_NoMatch(t *testing.T) {
	stream := NewRecordStream()
	rec := Events().ValueType(ValueTypeJob).FindFirst(stream)
	assert.Nil(t, rec)
}

func TestEventFilter_Count(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 1, "myJob", "task1", 100, 3),
		makeJobRecord("CREATED", 2, "myJob", "task2", 100, 3),
		makeJobRecord("COMPLETED", 3, "myJob", "task1", 100, 3),
	)

	assert.Equal(t, 2, Events().ValueType(ValueTypeJob).Intent(IntentCreated).Count(stream))
	assert.Equal(t, 1, Events().ValueType(ValueTypeJob).Intent(IntentCompleted).Count(stream))
	assert.Equal(t, 0, Events().ValueType(ValueTypeIncident).Count(stream))
}

// --- Predicate isolation ---

func TestEventFilter_PredicateIsolation(t *testing.T) {
	// Ensure that building a predicate doesn't affect the filter
	filter := Events().ValueType(ValueTypeJob)
	pred1 := filter.RecordPredicate()

	// Add more filters after getting pred1
	filter.Intent(IntentCreated)
	pred2 := filter.RecordPredicate()

	completedJob := makeJobRecord("COMPLETED", 1, "myJob", "task1", 100, 3)

	// pred1 should still match (only ValueType filter)
	assert.True(t, pred1(completedJob))
	// pred2 should NOT match (ValueType + Intent filter)
	assert.False(t, pred2(completedJob))
}
