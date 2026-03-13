package behavior

import (
	"context"
	"encoding/json"
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

const bpmnAdHocSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="adhoc-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:adHocSubProcess id="adhoc">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:extensionElements>
        <zeebe:adHoc activeElementsCollection="activateElements"/>
      </bpmn:extensionElements>
      <bpmn:serviceTask id="taskA" name="Task A">
        <bpmn:extensionElements>
          <zeebe:taskDefinition type="typeA"/>
        </bpmn:extensionElements>
      </bpmn:serviceTask>
      <bpmn:serviceTask id="taskB" name="Task B">
        <bpmn:extensionElements>
          <zeebe:taskDefinition type="typeB"/>
        </bpmn:extensionElements>
      </bpmn:serviceTask>
      <bpmn:task id="taskC" name="Task C"/>
    </bpmn:adHocSubProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="adhoc"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="adhoc" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func seedAdHocProcess(t *testing.T, store storage.Store) uint64 {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnAdHocSubProcess)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	var pdKey uint64 = 1
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           pdKey,
		BpmnProcessId: "adhoc-process",
		Version:       1,
		Content:       content,
		ContentHash:   []byte("adhoc-hash"),
		DeployedAt:    time.Now(),
	}))
	return pdKey
}

func TestAdHocSubProcess_ActivatesRequestedElements(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedAdHocProcess(t, store)

	ctx := context.Background()

	// Create process instance
	piKey := uint64(10)
	require.NoError(t, store.ProcessInstances().CreateInstance(ctx, &storage.ProcessInstance{
		Key:                  piKey,
		ProcessDefinitionKey: pdKey,
		BpmnProcessId:        "adhoc-process",
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}))

	// Set variable activateElements = ["taskA", "taskC"]
	varJSON, _ := json.Marshal([]string{"taskA", "taskC"})
	require.NoError(t, store.Variables().Create(ctx, &storage.Variable{
		Key:                1000,
		ProcessInstanceKey: piKey,
		ScopeKey:           piKey,
		Name:               "activateElements",
		Value:              varJSON,
	}))

	// Activate the ad-hoc subprocess
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: piKey},
		ProcessDefinitionKey: pdKey,
		ElementId:            "adhoc",
		ElementType:          "adHocSubProcess",
		FlowScopeKey:         piKey,
	})
	require.Len(t, intents, 2, "should activate taskA and taskC")

	// Verify first is taskA
	a1, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "taskA", a1.ElementId)
	assert.Equal(t, "serviceTask", a1.ElementType)
	assert.Equal(t, uint64(100), a1.FlowScopeKey, "ad-hoc subprocess key is the scope")
	assert.Equal(t, "typeA", a1.JobType)

	// Verify second is taskC (task, auto-complete)
	a2, ok := intents[1].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "taskC", a2.ElementId)
	assert.Equal(t, uint64(100), a2.FlowScopeKey)
}

func TestAdHocSubProcess_EmptyList_StaysActivated(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedAdHocProcess(t, store)

	ctx := context.Background()

	piKey := uint64(10)
	require.NoError(t, store.ProcessInstances().CreateInstance(ctx, &storage.ProcessInstance{
		Key:                  piKey,
		ProcessDefinitionKey: pdKey,
		BpmnProcessId:        "adhoc-process",
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}))

	varJSON, _ := json.Marshal([]string{})
	require.NoError(t, store.Variables().Create(ctx, &storage.Variable{
		Key:                1000,
		ProcessInstanceKey: piKey,
		ScopeKey:           piKey,
		Name:               "activateElements",
		Value:              varJSON,
	}))

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: piKey},
		ProcessDefinitionKey: pdKey,
		ElementId:            "adhoc",
		ElementType:          "adHocSubProcess",
		FlowScopeKey:         piKey,
	})
	assert.Empty(t, intents, "empty list → no elements activated, subprocess stays activated")
}

func TestAdHocSubProcess_ChildCompletion_CompletesSubProcess(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedAdHocProcess(t, store)

	ctx := context.Background()

	piKey := uint64(10)
	require.NoError(t, store.ProcessInstances().CreateInstance(ctx, &storage.ProcessInstance{
		Key:                  piKey,
		ProcessDefinitionKey: pdKey,
		BpmnProcessId:        "adhoc-process",
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}))

	// Create ad-hoc subprocess element instance (activated)
	adhocKey := uint64(100)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  adhocKey,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "adhoc",
		ElementType:          "adHocSubProcess",
		FlowScopeKey:         piKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Create child taskA element instance (already completed)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  200,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "taskA",
		ElementType:          "serviceTask",
		FlowScopeKey:         adhocKey,
		State:                storage.ElementInstanceCompleted,
		CreatedAt:            time.Now(),
	}))

	// Create child taskB element instance (activated — about to complete)
	taskBKey := uint64(300)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  taskBKey,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "taskB",
		ElementType:          "serviceTask",
		FlowScopeKey:         adhocKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Complete taskB — should trigger ad-hoc subprocess completion
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: taskBKey,
	})
	require.Len(t, intents, 1)

	// Should complete the ad-hoc subprocess
	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, adhocKey, complete.ElementInstanceKey)
}

func TestAdHocSubProcess_ChildCompletion_WaitsIfSiblingsActive(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedAdHocProcess(t, store)

	ctx := context.Background()

	piKey := uint64(10)
	require.NoError(t, store.ProcessInstances().CreateInstance(ctx, &storage.ProcessInstance{
		Key:                  piKey,
		ProcessDefinitionKey: pdKey,
		BpmnProcessId:        "adhoc-process",
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}))

	adhocKey := uint64(100)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  adhocKey,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "adhoc",
		ElementType:          "adHocSubProcess",
		FlowScopeKey:         piKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// taskA completing
	taskAKey := uint64(200)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  taskAKey,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "taskA",
		ElementType:          "serviceTask",
		FlowScopeKey:         adhocKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// taskB still active
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key:                  300,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: pdKey,
		ElementId:            "taskB",
		ElementType:          "serviceTask",
		FlowScopeKey:         adhocKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Complete taskA — should NOT complete ad-hoc (taskB still active)
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: taskAKey,
	})
	assert.Empty(t, intents, "should not complete ad-hoc subprocess while taskB is still active")
}
