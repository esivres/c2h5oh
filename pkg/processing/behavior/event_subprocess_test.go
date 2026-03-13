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

// Process with an interrupting error event subprocess at process level.
const bpmnErrorEventSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="mainSub">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:startEvent id="mainSubStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:endEvent id="errorEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Error_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="mainSubStart" targetRef="errorEnd"/>
    </bpmn:subProcess>
    <bpmn:subProcess id="errorHandler" triggeredByEvent="true">
      <bpmn:startEvent id="errorStartEvent" isInterrupting="true">
        <bpmn:outgoing>espFlow1</bpmn:outgoing>
        <bpmn:errorEventDefinition id="ErrEvtDef_2" errorRef="Error_1"/>
      </bpmn:startEvent>
      <bpmn:serviceTask id="handleTask" name="Handle Error">
        <bpmn:extensionElements>
          <zeebe:taskDefinition type="error-handler"/>
        </bpmn:extensionElements>
        <bpmn:incoming>espFlow1</bpmn:incoming>
        <bpmn:outgoing>espFlow2</bpmn:outgoing>
      </bpmn:serviceTask>
      <bpmn:endEvent id="espEnd">
        <bpmn:incoming>espFlow2</bpmn:incoming>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="espFlow1" sourceRef="errorStartEvent" targetRef="handleTask"/>
      <bpmn:sequenceFlow id="espFlow2" sourceRef="handleTask" targetRef="espEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="mainSub"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="mainSub" targetRef="end"/>
  </bpmn:process>
  <bpmn:error id="Error_1" name="BusinessError" errorCode="ERR_001"/>
</bpmn:definitions>`

// Non-interrupting error event subprocess.
const bpmnNonInterruptingErrorEventSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="mainSub">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:startEvent id="mainSubStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:endEvent id="errorEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Error_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="mainSubStart" targetRef="errorEnd"/>
    </bpmn:subProcess>
    <bpmn:subProcess id="errorHandler" triggeredByEvent="true">
      <bpmn:startEvent id="errorStartEvent" isInterrupting="false">
        <bpmn:outgoing>espFlow1</bpmn:outgoing>
        <bpmn:errorEventDefinition id="ErrEvtDef_2" errorRef="Error_1"/>
      </bpmn:startEvent>
      <bpmn:endEvent id="espEnd">
        <bpmn:incoming>espFlow1</bpmn:incoming>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="espFlow1" sourceRef="errorStartEvent" targetRef="espEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="mainSub"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="mainSub" targetRef="end"/>
  </bpmn:process>
  <bpmn:error id="Error_1" name="BusinessError" errorCode="ERR_001"/>
</bpmn:definitions>`

func seedEventSubProcess(t *testing.T, store storage.Store, bpmnXML string) uint64 {
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

func TestErrorEventSubProcess_Interrupting(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedEventSubProcess(t, store, bpmnErrorEventSubProcess)
	ctx := context.Background()
	piKey := uint64(10)

	// mainSub is active at process level scope
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 100, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "mainSub", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Error end event inside mainSub
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "errorEnd", ElementType: "endEvent",
		FlowScopeKey: 100, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the error end event — error should propagate to process level
	// mainSub has no boundary error event, so it should find the error event subprocess
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Interrupting event subprocess: should terminate mainSub + activate errorHandler
	var hasTerminateSub, hasActivateESP bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.TerminateElementIntent:
			if v.ElementInstanceKey == 100 {
				hasTerminateSub = true
			}
		case *intent.ActivateElementIntent:
			if v.ElementId == "errorHandler" && v.ElementType == "subProcess" {
				hasActivateESP = true
				assert.Equal(t, piKey, v.FlowScopeKey, "event subprocess scope should be process level")
			}
		}
	}
	assert.True(t, hasTerminateSub, "should terminate mainSub")
	assert.True(t, hasActivateESP, "should activate error event subprocess")
}

func TestErrorEventSubProcess_NonInterrupting(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedEventSubProcess(t, store, bpmnNonInterruptingErrorEventSubProcess)
	ctx := context.Background()
	piKey := uint64(10)

	// mainSub is active
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 100, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "mainSub", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Error end event inside mainSub
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "errorEnd", ElementType: "endEvent",
		FlowScopeKey: 100, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Non-interrupting: should NOT terminate mainSub, only activate event subprocess
	var hasTerminateSub, hasActivateESP bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.TerminateElementIntent:
			if v.ElementInstanceKey == 100 {
				hasTerminateSub = true
			}
		case *intent.ActivateElementIntent:
			if v.ElementId == "errorHandler" {
				hasActivateESP = true
			}
		}
	}
	assert.False(t, hasTerminateSub, "non-interrupting should NOT terminate mainSub")
	assert.True(t, hasActivateESP, "should activate error event subprocess")
}
