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

// Process with parallel gateway: two branches, one ends with terminate end event.
// start → parallelFork → [taskA, taskB → terminateEnd]
const bpmnWithTerminateEndEvent = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:parallelGateway id="fork">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:outgoing>flow3</bpmn:outgoing>
    </bpmn:parallelGateway>
    <bpmn:serviceTask id="taskA" name="Task A">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="worker-a"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:serviceTask id="taskB" name="Task B">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="worker-b"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow3</bpmn:incoming>
      <bpmn:outgoing>flow5</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="terminateEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
      <bpmn:terminateEventDefinition id="TermEvtDef_1"/>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="fork"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="fork" targetRef="taskA"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="fork" targetRef="taskB"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="taskA" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="taskB" targetRef="terminateEnd"/>
  </bpmn:process>
</bpmn:definitions>`

// Process with subprocess containing terminate end event.
const bpmnWithTerminateInSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="sub1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow6</bpmn:outgoing>
      <bpmn:startEvent id="subStart">
        <bpmn:outgoing>subFlow1</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:parallelGateway id="subFork">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:outgoing>subFlow2</bpmn:outgoing>
        <bpmn:outgoing>subFlow3</bpmn:outgoing>
      </bpmn:parallelGateway>
      <bpmn:serviceTask id="subTaskA" name="Sub Task A">
        <bpmn:extensionElements>
          <zeebe:taskDefinition type="sub-worker-a"/>
        </bpmn:extensionElements>
        <bpmn:incoming>subFlow2</bpmn:incoming>
        <bpmn:outgoing>subFlow4</bpmn:outgoing>
      </bpmn:serviceTask>
      <bpmn:endEvent id="subTermEnd">
        <bpmn:incoming>subFlow3</bpmn:incoming>
        <bpmn:terminateEventDefinition id="TermEvtDef_2"/>
      </bpmn:endEvent>
      <bpmn:endEvent id="subNormalEnd">
        <bpmn:incoming>subFlow4</bpmn:incoming>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="subFork"/>
      <bpmn:sequenceFlow id="subFlow2" sourceRef="subFork" targetRef="subTaskA"/>
      <bpmn:sequenceFlow id="subFlow3" sourceRef="subFork" targetRef="subTermEnd"/>
      <bpmn:sequenceFlow id="subFlow4" sourceRef="subTaskA" targetRef="subNormalEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow6</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow6" sourceRef="sub1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func seedTerminateProcess(t *testing.T, store storage.Store, bpmnXML string) uint64 {
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

func TestTerminateEndEvent_KillsParallelBranch(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedTerminateProcess(t, store, bpmnWithTerminateEndEvent)
	ctx := context.Background()

	piKey := uint64(10)

	// Create element instances: taskA is active in a parallel branch
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "taskA", ElementType: "serviceTask",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create terminate end event element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "terminateEnd", ElementType: "endEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the terminate end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should produce: TerminateElement(taskA) + CompleteProcessInstance
	require.Len(t, intents, 2)

	term, ok := intents[0].(*intent.TerminateElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(200), term.ElementInstanceKey)

	_, ok = intents[1].(*intent.CompleteProcessInstanceIntent)
	require.True(t, ok)
}

func TestTerminateEndEvent_InSubProcess_OnlyKillsSubProcessScope(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedTerminateProcess(t, store, bpmnWithTerminateInSubProcess)
	ctx := context.Background()

	piKey := uint64(10)
	subKey := uint64(100) // subprocess element instance key

	// Create subprocess element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: subKey, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "sub1", ElementType: "subProcess",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create subTaskA in subprocess scope — should be terminated
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "subTaskA", ElementType: "serviceTask",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create terminate end event in subprocess scope
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "subTermEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the terminate end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should produce: TerminateElement(subTaskA) + CompleteElement(sub1)
	// NOT CompleteProcessInstance — subprocess terminate only kills subprocess scope
	require.Len(t, intents, 2)

	term, ok := intents[0].(*intent.TerminateElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(200), term.ElementInstanceKey, "should terminate subTaskA")

	complete, ok := intents[1].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, subKey, complete.ElementInstanceKey, "should complete the subprocess")
}

func TestNormalEndEvent_DoesNotTerminateSiblings(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedTerminateProcess(t, store, bpmnWithTerminateEndEvent)
	ctx := context.Background()

	piKey := uint64(10)

	// Create element instances: taskB is active in a parallel branch
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "taskB", ElementType: "serviceTask",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Create normal end event element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: pdKey,
		ElementId: "normalEnd", ElementType: "endEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the normal end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Normal end event should just complete process, NOT terminate taskB
	require.Len(t, intents, 1)
	_, ok := intents[0].(*intent.CompleteProcessInstanceIntent)
	require.True(t, ok)
}
