package behavior

import (
	"context"
	"testing"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// Subprocess with error end event, caught by boundary error event on subprocess.
const bpmnErrorEndEventCaught = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="sub1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
      <bpmn:startEvent id="subStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:endEvent id="errorEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Error_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="errorEnd"/>
    </bpmn:subProcess>
    <bpmn:boundaryEvent id="errorBoundary" attachedToRef="sub1" cancelActivity="true">
      <bpmn:outgoing>flow5</bpmn:outgoing>
      <bpmn:errorEventDefinition id="ErrEvtDef_2" errorRef="Error_1"/>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="errorHandledEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="sub1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="errorBoundary" targetRef="errorHandledEnd"/>
  </bpmn:process>
  <bpmn:error id="Error_1" name="BusinessError" errorCode="ERR_001"/>
</bpmn:definitions>`

// Process-level error end event with no boundary event to catch it.
const bpmnErrorEndEventUncaught = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="sub1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:startEvent id="subStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:endEvent id="errorEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Error_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="errorEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="sub1" targetRef="end"/>
  </bpmn:process>
  <bpmn:error id="Error_1" name="BusinessError" errorCode="ERR_001"/>
</bpmn:definitions>`

// Subprocess with error end event, caught by catch-all boundary (no errorRef).
const bpmnErrorEndEventCatchAll = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="sub1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
      <bpmn:startEvent id="subStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:endEvent id="errorEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Error_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="errorEnd"/>
    </bpmn:subProcess>
    <bpmn:boundaryEvent id="catchAllBoundary" attachedToRef="sub1" cancelActivity="true">
      <bpmn:outgoing>flow5</bpmn:outgoing>
      <bpmn:errorEventDefinition id="ErrEvtDef_2"/>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="errorHandledEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="sub1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="catchAllBoundary" targetRef="errorHandledEnd"/>
  </bpmn:process>
  <bpmn:error id="Error_1" name="BusinessError" errorCode="ERR_001"/>
</bpmn:definitions>`

func seedErrorProcess(t *testing.T, store storage.Store, bpmnXML string) uint64 {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnXML)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	var pdKey uint64 = 1
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           pdKey,
		BpmnProcessId: "test-process",
		Version:       1,
		Content:       content,
		ContentHash:   []byte("test-hash"),
		DeployedAt:    time.Now(),
	}))
	return pdKey
}

func TestErrorEndEvent_CaughtByBoundary(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedErrorProcess(t, store, bpmnErrorEndEventCaught)
	ctx := context.Background()
	piKey := uint64(10)
	subKey := uint64(100)

	// Create subprocess element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: subKey, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "sub1", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create error end event element instance inside subprocess
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "errorEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the error end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should: TerminateElement(sub1) + ActivateElement(errorBoundary)
	var hasTerminateSub, hasActivateBoundary bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.TerminateElementIntent:
			if v.ElementInstanceKey == subKey {
				hasTerminateSub = true
			}
		case *intent.ActivateElementIntent:
			if v.ElementId == "errorBoundary" && v.ElementType == "boundaryEvent" {
				hasActivateBoundary = true
			}
		}
	}
	assert.True(t, hasTerminateSub, "should terminate subprocess")
	assert.True(t, hasActivateBoundary, "should activate error boundary event")
}

func TestErrorEndEvent_Uncaught_CreatesIncident(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedErrorProcess(t, store, bpmnErrorEndEventUncaught)
	ctx := context.Background()
	piKey := uint64(10)
	subKey := uint64(100)

	// Create subprocess element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: subKey, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "sub1", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create error end event
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "errorEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})
	require.Len(t, intents, 1)

	incident, ok := intents[0].(*intent.CreateIncidentIntent)
	require.True(t, ok)
	assert.Equal(t, "UNHANDLED_BPMN_ERROR", incident.ErrorType)
	assert.Contains(t, incident.ErrorMessage, "ERR_001")
}

func TestErrorEndEvent_CatchAllBoundary(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedErrorProcess(t, store, bpmnErrorEndEventCatchAll)
	ctx := context.Background()
	piKey := uint64(10)
	subKey := uint64(100)

	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: subKey, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "sub1", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "errorEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Catch-all boundary should be activated
	var hasActivateBoundary bool
	for _, in := range intents {
		if v, ok := in.(*intent.ActivateElementIntent); ok {
			if v.ElementId == "catchAllBoundary" {
				hasActivateBoundary = true
			}
		}
	}
	assert.True(t, hasActivateBoundary, "catch-all boundary should catch the error")
}
