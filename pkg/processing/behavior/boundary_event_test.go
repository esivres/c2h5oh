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

// Interrupting boundary timer on a service task.
const bpmnInterruptingBoundary = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="task1" name="Main Task">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="main-worker"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:boundaryEvent id="boundary1" attachedToRef="task1" cancelActivity="true">
      <bpmn:outgoing>flow3</bpmn:outgoing>
      <bpmn:timerEventDefinition id="TimerEvt_1">
        <bpmn:timeDuration>PT10S</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="timeoutEnd">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="boundary1" targetRef="timeoutEnd"/>
  </bpmn:process>
</bpmn:definitions>`

// Non-interrupting boundary timer on a service task.
const bpmnNonInterruptingBoundary = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="task1" name="Main Task">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="main-worker"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:boundaryEvent id="boundary1" attachedToRef="task1" cancelActivity="false">
      <bpmn:outgoing>flow3</bpmn:outgoing>
      <bpmn:timerEventDefinition id="TimerEvt_1">
        <bpmn:timeDuration>PT10S</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="timeoutEnd">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="boundary1" targetRef="timeoutEnd"/>
  </bpmn:process>
</bpmn:definitions>`

func seedBoundaryProcess(t *testing.T, store storage.Store, bpmnXML string) uint64 {
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

func TestBoundaryEvent_Interrupting_TerminatesAttachedElement(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedBoundaryProcess(t, store, bpmnInterruptingBoundary)
	ctx := context.Background()

	piKey := uint64(10)

	// Create attached task element instance (active)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "task1", ElementType: "serviceTask",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create boundary event element instance (activated by timer trigger)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "boundary1", ElementType: "boundaryEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the boundary event (timer fired)
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should produce: ActivateElement(timeoutEnd) + TerminateElement(task1)
	var hasActivate, hasTerminate bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.ActivateElementIntent:
			assert.Equal(t, "timeoutEnd", v.ElementId)
			hasActivate = true
		case *intent.TerminateElementIntent:
			assert.Equal(t, uint64(200), v.ElementInstanceKey)
			hasTerminate = true
		}
	}
	assert.True(t, hasActivate, "should activate timeoutEnd")
	assert.True(t, hasTerminate, "should terminate attached task1")
}

func TestBoundaryEvent_NonInterrupting_DoesNotTerminateAttachedElement(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedBoundaryProcess(t, store, bpmnNonInterruptingBoundary)
	ctx := context.Background()

	piKey := uint64(10)

	// Create attached task element instance (active)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "task1", ElementType: "serviceTask",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create non-interrupting boundary event element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "boundary1", ElementType: "boundaryEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the boundary event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should produce ONLY ActivateElement(timeoutEnd), NO TerminateElement
	require.Len(t, intents, 1)
	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "timeoutEnd", activate.ElementId)
}

func TestBoundaryEvent_DefaultCancelActivity_IsInterrupting(t *testing.T) {
	// cancelActivity defaults to true when not explicitly set
	bpmnDefaultCancel := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="task1" name="Main Task">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="main-worker"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:boundaryEvent id="boundary1" attachedToRef="task1">
      <bpmn:outgoing>flow3</bpmn:outgoing>
      <bpmn:timerEventDefinition id="TimerEvt_1">
        <bpmn:timeDuration>PT10S</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="timeoutEnd">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="boundary1" targetRef="timeoutEnd"/>
  </bpmn:process>
</bpmn:definitions>`

	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedBoundaryProcess(t, store, bpmnDefaultCancel)
	ctx := context.Background()
	piKey := uint64(10)

	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "task1", ElementType: "serviceTask",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "boundary1", ElementType: "boundaryEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Default cancelActivity=true → should terminate attached element
	var hasTerminate bool
	for _, in := range intents {
		if term, ok := in.(*intent.TerminateElementIntent); ok {
			assert.Equal(t, uint64(200), term.ElementInstanceKey)
			hasTerminate = true
		}
	}
	assert.True(t, hasTerminate, "default cancelActivity should be interrupting")
}
