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

// Process with signal intermediate throw and catch events.
const bpmnSignalThrowCatch = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="sender-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:intermediateThrowEvent id="signalThrow">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:signalEventDefinition id="SigEvtDef_1" signalRef="Signal_1"/>
    </bpmn:intermediateThrowEvent>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="signalThrow"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="signalThrow" targetRef="end"/>
  </bpmn:process>
  <bpmn:signal id="Signal_1" name="order-approved"/>
</bpmn:definitions>`

// Process with signal intermediate catch event.
const bpmnSignalCatch = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="receiver-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:intermediateCatchEvent id="signalCatch">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:signalEventDefinition id="SigEvtDef_1" signalRef="Signal_1"/>
    </bpmn:intermediateCatchEvent>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="signalCatch"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="signalCatch" targetRef="end"/>
  </bpmn:process>
  <bpmn:signal id="Signal_1" name="order-approved"/>
</bpmn:definitions>`

// Process with signal end event.
const bpmnSignalEndEvent = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="signal-end-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:endEvent id="signalEnd">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:signalEventDefinition id="SigEvtDef_1" signalRef="Signal_1"/>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="signalEnd"/>
  </bpmn:process>
  <bpmn:signal id="Signal_1" name="order-approved"/>
</bpmn:definitions>`

func seedSignalProcess(t *testing.T, store storage.Store, bpmnXML string, pdKey uint64, processId string) {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnXML)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           pdKey,
		BpmnProcessId: processId,
		Version:       1,
		Content:       content,
		ContentHash:   []byte("hash-" + processId),
		DeployedAt:    time.Now(),
	}))
}

func TestSignalCatchEvent_OpensSubscription(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedSignalProcess(t, store, bpmnSignalCatch, 1, "receiver-process")

	// Activate signal catch event
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "signalCatch",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	sub, ok := intents[0].(*intent.OpenSubscriptionIntent)
	require.True(t, ok)
	assert.Equal(t, "order-approved", sub.MessageName) // signal name stored as message name
	assert.Equal(t, "", sub.CorrelationKey)            // signals have empty correlation key
}

func TestSignalThrowEvent_EmitsThrowSignal(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedSignalProcess(t, store, bpmnSignalThrowCatch, 1, "sender-process")

	// Activate signal throw event
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "signalThrow",
		ElementType:          "intermediateThrowEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 2) // ThrowSignal + CompleteElement

	throwSig, ok := intents[0].(*intent.ThrowSignalIntent)
	require.True(t, ok)
	assert.Equal(t, "order-approved", throwSig.SignalName)

	complete, ok := intents[1].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), complete.ElementInstanceKey)
}

func TestThrowSignal_BroadcastsToAllSubscribers(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	// Create two open subscriptions for the same signal
	require.NoError(t, store.MessageSubscriptions().CreateSubscription(ctx, &storage.MessageSubscription{
		Key: 501, ProcessInstanceKey: 20, ElementInstanceKey: 201,
		MessageName: "order-approved", CorrelationKey: "",
		State: storage.MessageSubscriptionOpened, CreatedAt: time.Now(),
	}))
	require.NoError(t, store.MessageSubscriptions().CreateSubscription(ctx, &storage.MessageSubscription{
		Key: 502, ProcessInstanceKey: 30, ElementInstanceKey: 301,
		MessageName: "order-approved", CorrelationKey: "",
		State: storage.MessageSubscriptionOpened, CreatedAt: time.Now(),
	}))

	// Throw signal
	intents, err := store.Execute(ctx, AsHandler(
		Typed(throwSignal),
		&intent.ThrowSignalIntent{
			Header:     intent.Header{Key: 50},
			SignalName: "order-approved",
		},
	))
	require.NoError(t, err)

	// Both subscriptions should be correlated
	require.Len(t, intents, 2)
	for _, in := range intents {
		corr, ok := in.(*intent.CorrelateMessageIntent)
		require.True(t, ok)
		assert.Contains(t, []uint64{501, 502}, corr.SubscriptionKey)
	}
}

func TestSignalEndEvent_ThrowsSignal(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedSignalProcess(t, store, bpmnSignalEndEvent, 1, "signal-end-process")
	ctx := context.Background()
	piKey := uint64(10)

	// Create signal end event element instance
	require.NoError(t, store.ProcessInstances().CreateElementInstance(ctx, &storage.ElementInstance{
		Key: 300, ProcessInstanceKey: piKey, ProcessDefinitionKey: 1,
		ElementId: "signalEnd", ElementType: "endEvent",
		FlowScopeKey: piKey, State: storage.ElementInstanceActivated, CreatedAt: time.Now(),
	}))

	// Complete the signal end event
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: piKey},
		ElementInstanceKey: 300,
	})

	// Should: ThrowSignal + CompleteProcessInstance
	var hasThrow, hasComplete bool
	for _, in := range intents {
		switch v := in.(type) {
		case *intent.ThrowSignalIntent:
			assert.Equal(t, "order-approved", v.SignalName)
			hasThrow = true
		case *intent.CompleteProcessInstanceIntent:
			hasComplete = true
		}
	}
	assert.True(t, hasThrow, "should throw signal")
	assert.True(t, hasComplete, "should complete process")
}
