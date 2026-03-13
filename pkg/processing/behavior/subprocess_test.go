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

const bpmnWithSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:subProcess id="sub1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
      <bpmn:startEvent id="subStart">
        <bpmn:outgoing>flow2</bpmn:outgoing>
      </bpmn:startEvent>
      <bpmn:serviceTask id="subTask" name="Inner Task">
        <bpmn:extensionElements>
          <zeebe:taskDefinition type="inner-worker"/>
        </bpmn:extensionElements>
        <bpmn:incoming>flow2</bpmn:incoming>
        <bpmn:outgoing>flow3</bpmn:outgoing>
      </bpmn:serviceTask>
      <bpmn:endEvent id="subEnd">
        <bpmn:incoming>flow3</bpmn:incoming>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="flow2" sourceRef="subStart" targetRef="subTask"/>
      <bpmn:sequenceFlow id="flow3" sourceRef="subTask" targetRef="subEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="sub1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func seedSubProcess(t *testing.T, store storage.Store) uint64 {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnWithSubProcess)
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

func TestSubProcess_ActivatesInnerStartEvent(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedSubProcess(t, store)

	// Activate the subprocess
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "sub1",
		ElementType:          "subProcess",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "subStart", activate.ElementId)
	assert.Equal(t, "startEvent", activate.ElementType)
	// Inner elements use subprocess instance key as scope
	assert.Equal(t, uint64(100), activate.FlowScopeKey)
}

func TestSubProcess_InnerEndEvent_CompletesSubProcess(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedSubProcess(t, store)

	// Create subprocess element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  100,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: pdKey,
		ElementId:            "sub1",
		ElementType:          "subProcess",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Create inner end event element instance (scope = subprocess key 100)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  300,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: pdKey,
		ElementId:            "subEnd",
		ElementType:          "endEvent",
		FlowScopeKey:         100, // subprocess is the scope
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Complete the inner end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 300,
	})
	require.Len(t, intents, 1)

	// Should complete the subprocess (not the whole process instance)
	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), complete.ElementInstanceKey) // subprocess key
}
