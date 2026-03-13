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

// Process with link throw → link catch (same name "jump1").
const bpmnLinkEvents = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:intermediateThrowEvent id="linkThrow">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:linkEventDefinition id="LinkEvt_1" name="jump1"/>
    </bpmn:intermediateThrowEvent>
    <bpmn:intermediateCatchEvent id="linkCatch">
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:linkEventDefinition id="LinkEvt_2" name="jump1"/>
    </bpmn:intermediateCatchEvent>
    <bpmn:serviceTask id="task1" name="After Link">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="after-link"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:outgoing>flow3</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="linkThrow"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="linkCatch" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="task1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func TestLinkEvent_ThrowTeleportsToCatch(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	bmi, err := bpmn_model.ReadFromString(bpmnLinkEvents)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	piKey := uint64(10)

	// Create link throw element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 200, ProcessInstanceKey: piKey, ProcessDefinitionKey: 1,
		ElementId: "linkThrow", ElementType: "intermediateThrowEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete link throw → should teleport to linkCatch
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 200,
	})
	require.Len(t, intents, 1)

	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "linkCatch", activate.ElementId)
	assert.Equal(t, "intermediateCatchEvent", activate.ElementType)
}

func TestLinkCatchEvent_AutoCompletes(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	bmi, err := bpmn_model.ReadFromString(bpmnLinkEvents)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	// Activate link catch event → should auto-complete (not wait)
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "linkCatch",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), complete.ElementInstanceKey)
}
