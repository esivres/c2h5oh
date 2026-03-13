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

// Subprocess with escalation end event, caught by boundary escalation event.
const bpmnEscalationCaught = `<?xml version="1.0" encoding="UTF-8"?>
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
      <bpmn:endEvent id="escalationEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:escalationEventDefinition id="EscEvtDef_1" escalationRef="Esc_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="escalationEnd"/>
    </bpmn:subProcess>
    <bpmn:boundaryEvent id="escBoundary" attachedToRef="sub1" cancelActivity="false">
      <bpmn:outgoing>flow5</bpmn:outgoing>
      <bpmn:escalationEventDefinition id="EscEvtDef_2" escalationRef="Esc_1"/>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="escHandledEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="sub1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="escBoundary" targetRef="escHandledEnd"/>
  </bpmn:process>
  <bpmn:escalation id="Esc_1" name="Warning" escalationCode="ESC_001"/>
</bpmn:definitions>`

// Catch-all boundary escalation.
const bpmnEscalationCatchAll = `<?xml version="1.0" encoding="UTF-8"?>
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
      <bpmn:endEvent id="escalationEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:escalationEventDefinition id="EscEvtDef_1" escalationRef="Esc_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="escalationEnd"/>
    </bpmn:subProcess>
    <bpmn:boundaryEvent id="catchAllEscBoundary" attachedToRef="sub1" cancelActivity="false">
      <bpmn:outgoing>flow5</bpmn:outgoing>
      <bpmn:escalationEventDefinition id="EscEvtDef_2"/>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="escHandledEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="sub1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="catchAllEscBoundary" targetRef="escHandledEnd"/>
  </bpmn:process>
  <bpmn:escalation id="Esc_1" name="Warning" escalationCode="ESC_001"/>
</bpmn:definitions>`

// Escalation end event without handler.
const bpmnEscalationUncaught = `<?xml version="1.0" encoding="UTF-8"?>
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
      <bpmn:endEvent id="escalationEnd">
        <bpmn:incoming>subFlow1</bpmn:incoming>
        <bpmn:escalationEventDefinition id="EscEvtDef_1" escalationRef="Esc_1"/>
      </bpmn:endEvent>
      <bpmn:sequenceFlow id="subFlow1" sourceRef="subStart" targetRef="escalationEnd"/>
    </bpmn:subProcess>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="sub1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="sub1" targetRef="end"/>
  </bpmn:process>
  <bpmn:escalation id="Esc_1" name="Warning" escalationCode="ESC_001"/>
</bpmn:definitions>`

func seedEscalationProcess(t *testing.T, store storage.Store, bpmnXML string) uint64 {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnXML)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	var pdKey uint64 = 1
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key: pdKey, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))
	return pdKey
}

func TestEscalation_CaughtByBoundary(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedEscalationProcess(t, store, bpmnEscalationCaught)
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
		ElementId: "escalationEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	var hasActivateBoundary bool
	for _, in := range intents {
		if v, ok := in.(*intent.ActivateElementIntent); ok {
			if v.ElementId == "escBoundary" && v.ElementType == "boundaryEvent" {
				hasActivateBoundary = true
			}
		}
	}
	assert.True(t, hasActivateBoundary, "should activate escalation boundary event")
}

func TestEscalation_CatchAllBoundary(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedEscalationProcess(t, store, bpmnEscalationCatchAll)
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
		ElementId: "escalationEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	var hasActivateBoundary bool
	for _, in := range intents {
		if v, ok := in.(*intent.ActivateElementIntent); ok {
			if v.ElementId == "catchAllEscBoundary" {
				hasActivateBoundary = true
			}
		}
	}
	assert.True(t, hasActivateBoundary, "catch-all boundary should catch escalation")
}

func TestEscalation_Uncaught_CreatesIncident(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	pdKey := seedEscalationProcess(t, store, bpmnEscalationUncaught)
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
		ElementId: "escalationEnd", ElementType: "endEvent",
		FlowScopeKey: subKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})
	require.Len(t, intents, 1)

	incident, ok := intents[0].(*intent.CreateIncidentIntent)
	require.True(t, ok)
	assert.Equal(t, "UNHANDLED_ESCALATION", incident.ErrorType)
	assert.Contains(t, incident.ErrorMessage, "ESC_001")
}
