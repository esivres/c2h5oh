package bpmn_asserts

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ProcessEventInspections ---

func TestFindProcessEvents_FindFirst(t *testing.T) {
	stream := NewRecordStream()

	// Create PROCESS_INSTANCE ELEMENT_ACTIVATED with PROCESS type
	rec1 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey:   100,
		BpmnElementType:      "PROCESS",
		BpmnProcessID:        "processA",
		ProcessDefinitionKey: 50,
		ElementID:            "processA",
	})
	rec2 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 2, ProcessInstanceValue{
		ProcessInstanceKey:   200,
		BpmnElementType:      "PROCESS",
		BpmnProcessID:        "processA",
		ProcessDefinitionKey: 50,
		ElementID:            "processA",
	})
	feedRecords(stream, rec1, rec2)

	first := FindProcessEvents(stream).FindFirstProcessInstance()
	require.NotNil(t, first)
	assert.Equal(t, int64(100), first.ProcessInstanceKey)

	last := FindProcessEvents(stream).FindLastProcessInstance()
	require.NotNil(t, last)
	assert.Equal(t, int64(200), last.ProcessInstanceKey)
}

func TestFindProcessEvents_ByProcessDefinitionKey(t *testing.T) {
	stream := NewRecordStream()

	rec1 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey:   100,
		BpmnElementType:      "PROCESS",
		ProcessDefinitionKey: 50,
	})
	rec2 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 2, ProcessInstanceValue{
		ProcessInstanceKey:   200,
		BpmnElementType:      "PROCESS",
		ProcessDefinitionKey: 60,
	})
	feedRecords(stream, rec1, rec2)

	result := FindProcessEvents(stream).
		WithProcessDefinitionKey(60).
		FindFirstProcessInstance()
	require.NotNil(t, result)
	assert.Equal(t, int64(200), result.ProcessInstanceKey)
}

func TestFindProcessEvents_TriggeredByTimer(t *testing.T) {
	stream := NewRecordStream()

	timerRec := makeTimerRecord("TRIGGERED", 1, "timerStart", 100, 50)
	feedRecords(stream, timerRec)

	result := FindProcessEvents(stream).
		TriggeredByTimer("timerStart").
		FindFirstProcessInstance()
	require.NotNil(t, result)
	assert.Equal(t, int64(100), result.ProcessInstanceKey)
}

func TestFindProcessEvents_TriggeredByTimer_NoMatch(t *testing.T) {
	stream := NewRecordStream()

	timerRec := makeTimerRecord("TRIGGERED", 1, "otherTimer", 100, 50)
	feedRecords(stream, timerRec)

	result := FindProcessEvents(stream).
		TriggeredByTimer("timerStart").
		FindFirstProcessInstance()
	assert.Nil(t, result)
}

func TestFindProcessEvents_FindProcessInstance_ByIndex(t *testing.T) {
	stream := NewRecordStream()

	for i := int64(0); i < 3; i++ {
		rec := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", i+1, ProcessInstanceValue{
			ProcessInstanceKey: 100 + i,
			BpmnElementType:    "PROCESS",
		})
		feedRecords(stream, rec)
	}

	result := FindProcessEvents(stream).FindProcessInstance(1)
	require.NotNil(t, result)
	assert.Equal(t, int64(101), result.ProcessInstanceKey)

	outOfRange := FindProcessEvents(stream).FindProcessInstance(10)
	assert.Nil(t, outOfRange)
}

func TestFindProcessEvents_Empty(t *testing.T) {
	stream := NewRecordStream()

	assert.Nil(t, FindProcessEvents(stream).FindFirstProcessInstance())
	assert.Nil(t, FindProcessEvents(stream).FindLastProcessInstance())
}

// --- ProcessInstanceInspections ---

func TestFindProcessInstances_ByParent(t *testing.T) {
	stream := NewRecordStream()

	child := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey:       200,
		ParentProcessInstanceKey: 100,
		BpmnElementType:          "PROCESS",
		BpmnProcessID:            "childProcess",
	})
	unrelated := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 2, ProcessInstanceValue{
		ProcessInstanceKey:       300,
		ParentProcessInstanceKey: 999,
		BpmnElementType:          "PROCESS",
		BpmnProcessID:            "otherProcess",
	})
	feedRecords(stream, child, unrelated)

	result := FindProcessInstances(stream).
		WithParentProcessInstanceKey(100).
		FindFirstProcessInstance()
	require.NotNil(t, result)
	assert.Equal(t, int64(200), result.ProcessInstanceKey)
}

func TestFindProcessInstances_ByBpmnProcessID(t *testing.T) {
	stream := NewRecordStream()

	rec1 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey: 100,
		BpmnElementType:    "PROCESS",
		BpmnProcessID:      "processA",
	})
	rec2 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 2, ProcessInstanceValue{
		ProcessInstanceKey: 200,
		BpmnElementType:    "PROCESS",
		BpmnProcessID:      "processB",
	})
	feedRecords(stream, rec1, rec2)

	result := FindProcessInstances(stream).
		WithBpmnProcessID("processB").
		FindFirstProcessInstance()
	require.NotNil(t, result)
	assert.Equal(t, int64(200), result.ProcessInstanceKey)
}

func TestFindProcessInstances_CombinedFilters(t *testing.T) {
	stream := NewRecordStream()

	rec := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey:       200,
		ParentProcessInstanceKey: 100,
		BpmnElementType:          "PROCESS",
		BpmnProcessID:            "childProcess",
	})
	feedRecords(stream, rec)

	result := FindProcessInstances(stream).
		WithParentProcessInstanceKey(100).
		WithBpmnProcessID("childProcess").
		FindFirstProcessInstance()
	require.NotNil(t, result)
	assert.Equal(t, int64(200), result.ProcessInstanceKey)

	// Wrong parent
	noMatch := FindProcessInstances(stream).
		WithParentProcessInstanceKey(999).
		WithBpmnProcessID("childProcess").
		FindFirstProcessInstance()
	assert.Nil(t, noMatch)
}

// --- InspectedProcessInstance.AssertThat ---

func TestInspectedProcessInstance_AssertThat(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "testProcess", "PROCESS"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS"),
	)

	inspected := &InspectedProcessInstance{ProcessInstanceKey: 100}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	inspected.AssertThat(t, stream).WithContext(ctx).IsCompleted()
}

// --- Logger ---

func TestPrintCompact(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "start", "START_EVENT"),
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
		makeVariableRecord("CREATED", 3, 100, "result", `"ok"`),
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "no retries", "task1", 10),
	)

	output := stream.PrintCompact()
	assert.Contains(t, output, "PROCESS_INSTANCE")
	assert.Contains(t, output, "JOB")
	assert.Contains(t, output, "VARIABLE")
	assert.Contains(t, output, "INCIDENT")
	assert.Contains(t, output, "Unresolved incident")
	assert.Contains(t, output, "JOB_NO_RETRIES")
}

func TestPrintCompact_NoIncidents(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT"),
	)

	output := stream.PrintCompact()
	assert.NotContains(t, output, "Unresolved incident")
}

func TestPrintCompact_ResolvedIncidents(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
		makeIncidentRecord("RESOLVED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
	)

	output := stream.PrintCompact()
	assert.NotContains(t, output, "Unresolved incident")
}

func TestPrintCompact_DeploymentRecord(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeDeploymentRecord("CREATED", 1, []string{"myProcess"}, []string{"proc.bpmn"}),
	)

	output := stream.PrintCompact()
	assert.Contains(t, output, "DEPLOYMENT")
	assert.Contains(t, output, "proc.bpmn")
}

func TestPrintCompact_MessageRecords(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageRecord("PUBLISHED", 1, "paymentReceived", "corr-123"),
		makeMessageSubscriptionRecord("CREATED", 2, 100, 50, "paymentReceived"),
	)

	output := stream.PrintCompact()
	assert.Contains(t, output, "MESSAGE")
	assert.Contains(t, output, "paymentReceived")
}

// Helper to build raw JSON for feedRecords from Record
func init() {
	// verify feedRecords works correctly - sanity check
	stream := NewRecordStream()
	rec := makeProcessInstanceRecord("ELEMENT_ACTIVATED", 1, 100, "start", "START_EVENT")
	raw, _ := json.Marshal(rec)
	stream.Add(raw)

	records := stream.Filter(func(r Record) bool { return true })
	if len(records) != 1 {
		panic("feedRecords sanity check failed")
	}
	if records[0].ValueType != "PROCESS_INSTANCE" {
		panic("feedRecords sanity check: wrong valueType")
	}
}
