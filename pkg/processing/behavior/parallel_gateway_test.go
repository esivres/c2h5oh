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

// Build: start → fork(parallel) → [task1, task2] → join(parallel) → end
const bpmnParallelForkJoin = `<?xml version="1.0" encoding="UTF-8"?>
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
    <bpmn:serviceTask id="task1" name="Task 1">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="worker1"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:serviceTask id="task2" name="Task 2">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="worker2"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow3</bpmn:incoming>
      <bpmn:outgoing>flow5</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:parallelGateway id="join">
      <bpmn:incoming>flow4</bpmn:incoming>
      <bpmn:incoming>flow5</bpmn:incoming>
      <bpmn:outgoing>flow6</bpmn:outgoing>
    </bpmn:parallelGateway>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow6</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="fork"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="fork" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="fork" targetRef="task2"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="task1" targetRef="join"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="task2" targetRef="join"/>
    <bpmn:sequenceFlow id="flow6" sourceRef="join" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func seedParallelProcess(t *testing.T, store storage.Store) uint64 {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnParallelForkJoin)
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

func TestParallelGateway_Fork(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedParallelProcess(t, store)

	// Create element instance for the fork gateway
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  100,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: pdKey,
		ElementId:            "fork",
		ElementType:          "parallelGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Complete fork gateway → should activate both outgoing flows
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 2)

	elementIds := map[string]bool{}
	for _, i := range intents {
		act, ok := i.(*intent.ActivateElementIntent)
		require.True(t, ok)
		elementIds[act.ElementId] = true
	}
	assert.True(t, elementIds["task1"])
	assert.True(t, elementIds["task2"])
}

func TestParallelGateway_Join_FirstToken_Waits(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedParallelProcess(t, store)

	// First token arrives at join gateway
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 200, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "join",
		ElementType:          "parallelGateway",
		FlowScopeKey:         10,
	})
	// Should NOT complete — still waiting for second token
	assert.Empty(t, intents)
}

func TestParallelGateway_Join_AllTokens_Completes(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedParallelProcess(t, store)

	// First token arrives — wait
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 200, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "join",
		ElementType:          "parallelGateway",
		FlowScopeKey:         10,
	})
	assert.Empty(t, intents)

	// Second token arrives — should complete
	intents = executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 201, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "join",
		ElementType:          "parallelGateway",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(201), complete.ElementInstanceKey)
}

func TestParallelGateway_SingleIncoming_AutoCompletes(t *testing.T) {
	// Fork gateway has only 1 incoming flow → auto-complete (no join needed)
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedParallelProcess(t, store)

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 200, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "fork",
		ElementType:          "parallelGateway",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	_, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
}
