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

const bpmnWithTimerCatch = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:intermediateCatchEvent id="timer1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:timerEventDefinition id="timerDef1">
        <bpmn:timeDuration>PT30S</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:intermediateCatchEvent>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="timer1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="timer1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

const bpmnWithMessageCatch = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:message id="Msg_1" name="order-received"/>
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:intermediateCatchEvent id="msgCatch1">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:messageEventDefinition id="msgDef1" messageRef="Msg_1"/>
    </bpmn:intermediateCatchEvent>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="msgCatch1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="msgCatch1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func seedBpmnFromXML(t *testing.T, store storage.Store, pdKey uint64, bpmnXml string) {
	t.Helper()
	bmi, err := bpmn_model.ReadFromString(bpmnXml)
	require.NoError(t, err)
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           pdKey,
		BpmnProcessId: "test-process",
		Version:       1,
		Content:       content,
		ContentHash:   []byte("test-hash"),
		DeployedAt:    time.Now(),
	}))
}

func TestIntermediateCatchEvent_Timer_CreatesTimer(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	var pdKey uint64 = 1
	seedBpmnFromXML(t, store, pdKey, bpmnWithTimerCatch)

	before := time.Now()
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "timer1",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	timer, ok := intents[0].(*intent.CreateTimerIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), timer.ElementInstanceKey)
	// Duration is PT30S = 30 seconds from now
	assert.True(t, timer.DueDate.After(before.Add(29*time.Second)))
	assert.True(t, timer.DueDate.Before(before.Add(31*time.Second)))
}

func TestIntermediateCatchEvent_Message_OpensSubscription(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	var pdKey uint64 = 1
	seedBpmnFromXML(t, store, pdKey, bpmnWithMessageCatch)

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: pdKey,
		ElementId:            "msgCatch1",
		ElementType:          "intermediateCatchEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	sub, ok := intents[0].(*intent.OpenSubscriptionIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), sub.ElementInstanceKey)
	assert.Equal(t, "order-received", sub.MessageName)
}

func TestIntermediateThrowEvent_AutoCompletes(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 0,
		ElementId:            "throw1",
		ElementType:          "intermediateThrowEvent",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	_, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
}

func TestParseISO8601Duration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"PT10S", 10 * time.Second},
		{"PT1M", 1 * time.Minute},
		{"PT1H", 1 * time.Hour},
		{"PT1H30M", 90 * time.Minute},
		{"P1D", 24 * time.Hour},
		{"P1DT12H", 36 * time.Hour},
		{"PT0S", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := parseISO8601Duration(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, d)
		})
	}
}
