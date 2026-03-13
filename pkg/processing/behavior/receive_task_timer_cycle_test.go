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

// Process with receive task.
const bpmnReceiveTask = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:receiveTask id="receiveMsg" messageRef="Msg_1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
    </bpmn:receiveTask>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="receiveMsg"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="receiveMsg" targetRef="end"/>
  </bpmn:process>
  <bpmn:message id="Msg_1" name="order-received"/>
</bpmn:definitions>`

// Boundary timer with cycle.
const bpmnTimerCycleBoundary = `<?xml version="1.0" encoding="UTF-8"?>
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
    <bpmn:boundaryEvent id="timerBoundary" attachedToRef="task1" cancelActivity="false">
      <bpmn:outgoing>flow3</bpmn:outgoing>
      <bpmn:timerEventDefinition id="TimerEvt_1">
        <bpmn:timeCycle>R3/PT10S</bpmn:timeCycle>
      </bpmn:timerEventDefinition>
    </bpmn:boundaryEvent>
    <bpmn:endEvent id="normalEnd">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="timerEnd">
      <bpmn:incoming>flow3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="normalEnd"/>
    <bpmn:sequenceFlow id="flow3" sourceRef="timerBoundary" targetRef="timerEnd"/>
  </bpmn:process>
</bpmn:definitions>`

func TestReceiveTask_OpensSubscription(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	bmi, err := bpmn_model.ReadFromString(bpmnReceiveTask)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "receiveMsg",
		ElementType:          "receiveTask",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	sub, ok := intents[0].(*intent.OpenSubscriptionIntent)
	require.True(t, ok)
	assert.Equal(t, "order-received", sub.MessageName)
	assert.Equal(t, uint64(100), sub.ElementInstanceKey)
}

func TestTimerCycle_ParsesAndCreatesTimer(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	bmi, err := bpmn_model.ReadFromString(bpmnTimerCycleBoundary)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(ctx, &storage.ProcessDefinition{
		Key: 1, BpmnProcessId: "test-process", Version: 1,
		Content: content, ContentHash: []byte("h"), DeployedAt: time.Now(),
	}))

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "timerBoundary",
		ElementType:          "boundaryEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	timer, ok := intents[0].(*intent.CreateTimerIntent)
	require.True(t, ok)
	assert.Equal(t, 3, timer.Repetitions)
	assert.Equal(t, 10*time.Second, timer.CycleDuration)
}

func TestTimerCycle_TriggerCreatesNextTimer(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	// Create a cycle timer with 3 repetitions
	require.NoError(t, store.Timers().Create(ctx, &storage.Timer{
		Key: 500, ProcessInstanceKey: 10, ElementInstanceKey: 100,
		ProcessDefinitionKey: 1, State: storage.TimerCreated,
		DueDate: time.Now().Add(-time.Second), Repetitions: 3,
		CycleDuration: 10 * time.Second, CreatedAt: time.Now(),
	}))

	// Trigger the timer
	intents, err := store.Execute(ctx, AsHandler(
		Typed(triggerTimer),
		&intent.TriggerTimerIntent{
			Header:   intent.Header{Key: 50, ProcessInstanceKey: 10},
			TimerKey: 500,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 2) // CompleteElement + CreateTimer(next)

	_, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)

	nextTimer, ok := intents[1].(*intent.CreateTimerIntent)
	require.True(t, ok)
	assert.Equal(t, 2, nextTimer.Repetitions) // decremented
	assert.Equal(t, 10*time.Second, nextTimer.CycleDuration)
}

func TestTimerCycle_LastRepetition_NoNextTimer(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	ctx := context.Background()

	// Create a timer with 1 repetition left
	require.NoError(t, store.Timers().Create(ctx, &storage.Timer{
		Key: 500, ProcessInstanceKey: 10, ElementInstanceKey: 100,
		ProcessDefinitionKey: 1, State: storage.TimerCreated,
		DueDate: time.Now().Add(-time.Second), Repetitions: 1,
		CycleDuration: 10 * time.Second, CreatedAt: time.Now(),
	}))

	intents, err := store.Execute(ctx, AsHandler(
		Typed(triggerTimer),
		&intent.TriggerTimerIntent{
			Header:   intent.Header{Key: 50, ProcessInstanceKey: 10},
			TimerKey: 500,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1) // Only CompleteElement, no next timer

	_, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
}

func TestParseTimerCycle(t *testing.T) {
	tests := []struct {
		expr string
		reps int
		dur  time.Duration
	}{
		{"R3/PT10S", 3, 10 * time.Second},
		{"R/PT1M", -1, time.Minute},
		{"R1/PT5S", 1, 5 * time.Second},
		{"R10/PT1H", 10, time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			reps, dur, err := parseTimerCycle(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.reps, reps)
			assert.Equal(t, tt.dur, dur)
		})
	}
}
