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

// Process with compensation: task → boundary compensation event → handler task.
// Compensation end event triggers the compensation handler.
const bpmnCompensation = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="bookFlight" name="Book Flight">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="book-flight"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:boundaryEvent id="compBoundary" attachedToRef="bookFlight">
      <bpmn:compensateEventDefinition id="CompEvtDef_1"/>
    </bpmn:boundaryEvent>
    <bpmn:serviceTask id="cancelFlight" name="Cancel Flight" isForCompensation="true">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="cancel-flight"/>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:association id="assoc1" sourceRef="compBoundary" targetRef="cancelFlight"/>
    <bpmn:endEvent id="compensateEnd">
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:compensateEventDefinition id="CompEvtDef_2"/>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="bookFlight"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="bookFlight" targetRef="compensateEnd"/>
  </bpmn:process>
</bpmn:definitions>`

func TestCompensationEndEvent_ActivatesHandler(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	bmi, err := bpmn_model.ReadFromString(bpmnCompensation)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	piKey := uint64(10)

	// Create compensation end event element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: 1,
		ElementId: "compensateEnd", ElementType: "endEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should activate compensation handler + complete process
	var hasActivateHandler, hasCompleteProcess bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.ActivateElementIntent:
			if v.ElementId == "cancelFlight" {
				hasActivateHandler = true
				assert.Equal(t, "serviceTask", v.ElementType)
			}
		case *intent.CompleteProcessInstanceIntent:
			hasCompleteProcess = true
		}
	}
	assert.True(t, hasActivateHandler, "should activate compensation handler")
	assert.True(t, hasCompleteProcess, "should complete process instance")
}

func TestCompensationIntermediateThrow_ActivatesHandler(t *testing.T) {
	// Similar process but with intermediate throw compensation event
	bpmnCompThrow := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="bookFlight" name="Book Flight">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="book-flight"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:boundaryEvent id="compBoundary" attachedToRef="bookFlight">
      <bpmn:compensateEventDefinition id="CompEvtDef_1"/>
    </bpmn:boundaryEvent>
    <bpmn:serviceTask id="cancelFlight" name="Cancel Flight" isForCompensation="true">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="cancel-flight"/>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:association id="assoc1" sourceRef="compBoundary" targetRef="cancelFlight"/>
    <bpmn:intermediateThrowEvent id="compensateThrow">
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:outgoing>flow3</bpmn:outgoing>
      <bpmn:compensateEventDefinition id="CompEvtDef_2"/>
    </bpmn:intermediateThrowEvent>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="bookFlight"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="bookFlight" targetRef="compensateThrow"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="compensateThrow" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	bmi, err := bpmn_model.ReadFromString(bpmnCompThrow)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	piKey := uint64(10)

	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: 1,
		ElementId: "compensateThrow", ElementType: "intermediateThrowEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should activate compensation handler + continue to end via outgoing flow
	var hasActivateHandler, hasActivateEnd bool
	for _, in := range intents {
		if v, ok := in.(*intent.ActivateElementIntent); ok {
			if v.ElementId == "cancelFlight" {
				hasActivateHandler = true
			}
			if v.ElementId == "end" {
				hasActivateEnd = true
			}
		}
	}
	assert.True(t, hasActivateHandler, "should activate compensation handler")
	assert.True(t, hasActivateEnd, "should continue to end event")
}
