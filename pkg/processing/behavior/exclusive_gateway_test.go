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

func buildXorProcess(t *testing.T) *bpmn_model.BpmnModelInstance {
	t.Helper()
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ExclusiveGateway("xor1").
		ConditionExpression("= amount > 100").
		ServiceTask("bigOrder").ZeebeJobType("big-worker").
		EndEvent("end1").
		MoveToNode("xor1").
		DefaultFlow().
		ServiceTask("smallOrder").ZeebeJobType("small-worker").
		EndEvent("end2").
		Done()
	return bmi
}

func seedXorGateway(t *testing.T, store storage.Store, bmi *bpmn_model.BpmnModelInstance) uint64 {
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

	// Create element instance for the gateway
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  100,
		ProcessInstanceKey:   10,
		ProcessDefinitionKey: pdKey,
		ElementId:            "xor1",
		ElementType:          "exclusiveGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	return pdKey
}

func TestExclusiveGateway_ConditionTrue(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	bmi := buildXorProcess(t)
	seedXorGateway(t, store, bmi)

	// Set amount = 200 (condition "amount > 100" is true)
	amountJSON, _ := json.Marshal(200)
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "amount",
		Value:              amountJSON,
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)

	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "bigOrder", activate.ElementId)
	assert.Equal(t, "serviceTask", activate.ElementType)
	assert.Equal(t, "big-worker", activate.JobType)
}

func TestExclusiveGateway_DefaultFlow(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	bmi := buildXorProcess(t)
	seedXorGateway(t, store, bmi)

	// Set amount = 50 (condition "amount > 100" is false → default flow)
	amountJSON, _ := json.Marshal(50)
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "amount",
		Value:              amountJSON,
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)

	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "smallOrder", activate.ElementId)
}

func TestExclusiveGateway_NoMatchNoDefault_Incident(t *testing.T) {
	// Build a process where gateway has conditions but NO default flow
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ExclusiveGateway("xor1").
		ConditionExpression("= status = \"approved\"").
		ServiceTask("approved").ZeebeJobType("approve-worker").
		EndEvent("end1").
		MoveToNode("xor1").
		ConditionExpression("= status = \"rejected\"").
		ServiceTask("rejected").ZeebeJobType("reject-worker").
		EndEvent("end2").
		Done()

	store := sqlstore.NewStore(openTestDB(t))
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
		ElementId:            "xor1",
		ElementType:          "exclusiveGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Set status = "pending" — neither condition matches
	statusJSON, _ := json.Marshal("pending")
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "status",
		Value:              statusJSON,
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)

	inc, ok := intents[0].(*intent.CreateIncidentIntent)
	require.True(t, ok)
	assert.Equal(t, "CONDITION_ERROR", inc.ErrorType)
	assert.Contains(t, inc.ErrorMessage, "no outgoing flow taken")
}

func TestExclusiveGateway_SingleOutgoing_NoCondition(t *testing.T) {
	// Gateway with single outgoing flow — should pass through without condition
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ExclusiveGateway("xor1").
		EndEvent("end1").
		Done()

	store := sqlstore.NewStore(openTestDB(t))
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
		ElementId:            "xor1",
		ElementType:          "exclusiveGateway",
		FlowScopeKey:         10,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header:             intent.Header{Key: 50, ProcessInstanceKey: 10},
		ElementInstanceKey: 100,
	})
	require.Len(t, intents, 1)

	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "end1", activate.ElementId)
}
