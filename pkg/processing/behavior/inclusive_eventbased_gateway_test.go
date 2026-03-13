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

// --- Inclusive Gateway ---

func buildInclusiveProcess(t *testing.T) *bpmn_model.BpmnModelInstance {
	t.Helper()
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		InclusiveGateway("igw1").
		ConditionExpression("= needsReview").
		ServiceTask("review").ZeebeJobType("review-worker").
		EndEvent("end1").
		MoveToNode("igw1").
		ConditionExpression("= needsApproval").
		ServiceTask("approve").ZeebeJobType("approve-worker").
		EndEvent("end2").
		MoveToNode("igw1").
		DefaultFlow().
		ServiceTask("fallback").ZeebeJobType("fallback-worker").
		EndEvent("end3").
		Done()
	return bmi
}

func seedInclusiveGateway(t *testing.T, store storage.Store, bmi *bpmn_model.BpmnModelInstance) uint64 {
	t.Helper()
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

	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  100,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: pdKey,
		ElementId:            "igw1",
		ElementType:          "inclusiveGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))
	return pdKey
}

func setVar(t *testing.T, store storage.Store, key uint64, name string, value any) {
	t.Helper()
	v, _ := json.Marshal(value)
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                key,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               name,
		Value:              v,
	}))
}

func TestInclusiveGateway_BothConditionsTrue(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	bmi := buildInclusiveProcess(t)
	seedInclusiveGateway(t, store, bmi)

	setVar(t, store, 1, "needsReview", true)
	setVar(t, store, 2, "needsApproval", true)

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 2)

	ids := map[string]bool{}
	for _, i := range intents {
		ids[i.(*intent.ActivateElementIntent).ElementId] = true
	}
	assert.True(t, ids["review"])
	assert.True(t, ids["approve"])
}

func TestInclusiveGateway_OneConditionTrue(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	bmi := buildInclusiveProcess(t)
	seedInclusiveGateway(t, store, bmi)

	setVar(t, store, 1, "needsReview", true)
	setVar(t, store, 2, "needsApproval", false)

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)
	assert.Equal(t, "review", intents[0].(*intent.ActivateElementIntent).ElementId)
}

func TestInclusiveGateway_NoneTrue_DefaultFlow(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	bmi := buildInclusiveProcess(t)
	seedInclusiveGateway(t, store, bmi)

	setVar(t, store, 1, "needsReview", false)
	setVar(t, store, 2, "needsApproval", false)

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)
	assert.Equal(t, "fallback", intents[0].(*intent.ActivateElementIntent).ElementId)
}

// --- Event-Based Gateway ---

const bpmnWithEventBasedGw = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:message id="Msg_1" name="payment-received"/>
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:eventBasedGateway id="ebgw1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:outgoing>flow3</bpmn:outgoing>
    </bpmn:eventBasedGateway>
    <bpmn:intermediateCatchEvent id="timerCatch">
      <bpmn:incoming>flow2</bpmn:incoming>
      <bpmn:outgoing>flow4</bpmn:outgoing>
      <bpmn:timerEventDefinition id="timerDef1">
        <bpmn:timeDuration>PT1H</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:intermediateCatchEvent>
    <bpmn:intermediateCatchEvent id="msgCatch">
      <bpmn:incoming>flow3</bpmn:incoming>
      <bpmn:outgoing>flow5</bpmn:outgoing>
      <bpmn:messageEventDefinition id="msgDef1" messageRef="Msg_1"/>
    </bpmn:intermediateCatchEvent>
    <bpmn:endEvent id="timeoutEnd">
      <bpmn:incoming>flow4</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="successEnd">
      <bpmn:incoming>flow5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="ebgw1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="ebgw1" targetRef="timerCatch"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="ebgw1" targetRef="msgCatch"/>
    <bpmn:sequenceFlow id="flow4" sourceRef="timerCatch" targetRef="timeoutEnd"/>
    <bpmn:sequenceFlow id="flow5" sourceRef="msgCatch" targetRef="successEnd"/>
  </bpmn:process>
</bpmn:definitions>`

func TestEventBasedGateway_ActivatesBothCatchEvents(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedBpmnFromXML(t, store, 1, bpmnWithEventBasedGw)

	// Create gateway element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  100,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: 1,
		ElementId:            "ebgw1",
		ElementType:          "eventBasedGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Complete the gateway → should activate both catch events
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 2)

	ids := map[string]bool{}
	for _, i := range intents {
		ids[i.(*intent.ActivateElementIntent).ElementId] = true
	}
	assert.True(t, ids["timerCatch"])
	assert.True(t, ids["msgCatch"])
}

func TestEventBasedGateway_WinnerTerminatesSiblings(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedBpmnFromXML(t, store, 1, bpmnWithEventBasedGw)

	// Create both catch event element instances (as if gateway already completed)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  200,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: 1,
		ElementId:            "timerCatch",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  201,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: 1,
		ElementId:            "msgCatch",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Message catch event wins (completes)
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 201, // msgCatch
	})

	// Should have: ActivateElement(successEnd) + TerminateElement(timerCatch)
	var hasActivate, hasTerminate bool
	for _, i := range intents {
		switch v := i.(type) {
		case *intent.ActivateElementIntent:
			assert.Equal(t, "successEnd", v.ElementId)
			hasActivate = true
		case *intent.TerminateElementIntent:
			assert.Equal(t, uint64(200), v.ElementInstanceKey) // timerCatch
			hasTerminate = true
		}
	}
	assert.True(t, hasActivate, "should activate successEnd")
	assert.True(t, hasTerminate, "should terminate sibling timerCatch")
}
