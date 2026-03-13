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

func seedProcessDefinition(t *testing.T, s storage.Store, bmi *bpmn_model.BpmnModelInstance) *storage.ProcessDefinition {
	t.Helper()
	content, err := bmi.GetDocument().WriteToBytes()
	require.NoError(t, err)

	def := &storage.ProcessDefinition{
		Key:           1,
		BpmnProcessId: "test-process",
		Version:       1,
		Content:       content,
		ContentHash:   []byte("test-hash"),
		DeployedAt:    time.Now(),
	}
	require.NoError(t, s.ProcessDefinitions().Create(context.Background(), def))
	return def
}

func TestInputMapping_EvalAndCreateVariable(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Build process with input mapping on service task
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").ZeebeJobType("worker").
		ZeebeInput("= order.amount", "taskAmount").
		EndEvent("end").
		Done()

	def := seedProcessDefinition(t, store, bmi)

	// Create process instance and set variables in scope
	piKey := uint64(100)
	seedProcessInstance(t, store, piKey, storage.ProcessInstanceActive)

	// Set parent scope variables
	orderJson, _ := json.Marshal(map[string]any{"amount": 500, "id": "ORD-1"})
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: piKey,
		ScopeKey:           piKey,
		Name:               "order",
		Value:              orderJson,
	}))

	// Activate element (should apply input mappings)
	elementKey := uint64(201)
	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(activateElement),
		&intent.ActivateElementIntent{
			Header: intent.Header{
				Key:                elementKey,
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ProcessDefinitionKey: def.Key,
			ElementId:            "task1",
			ElementType:          "serviceTask",
			FlowScopeKey:         piKey,
			JobType:              "worker",
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1) // CreateJobIntent

	// Check that input mapping created variable in element scope
	v, err := store.Variables().FindByName(context.Background(), elementKey, "taskAmount")
	require.NoError(t, err)
	require.NotNil(t, v, "input mapping should create taskAmount variable")

	var val any
	require.NoError(t, json.Unmarshal(v.Value, &val))
	assert.InDelta(t, float64(500), val, 0.01) // JSON numbers are float64
}

func TestOutputMapping_EvalAndUpdateVariable(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Build process with output mapping on service task
	bmi := bpmn_model.CreateExecutableProcess("test-process").
		StartEvent("start").
		ServiceTask("task1").ZeebeJobType("worker").
		ZeebeOutput("= result", "taskResult").
		EndEvent("end").
		Done()

	def := seedProcessDefinition(t, store, bmi)

	piKey := uint64(100)
	seedProcessInstance(t, store, piKey, storage.ProcessInstanceActive)

	// Create element instance
	elementKey := uint64(201)
	require.NoError(t, store.ProcessInstances().CreateElementInstance(context.Background(), &storage.ElementInstance{
		Key:                  elementKey,
		ProcessInstanceKey:   piKey,
		ProcessDefinitionKey: def.Key,
		ElementId:            "task1",
		ElementType:          "serviceTask",
		FlowScopeKey:         piKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}))

	// Set variable in element scope (simulating worker output)
	resultJson, _ := json.Marshal("completed successfully")
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                2,
		ProcessInstanceKey: piKey,
		ScopeKey:           elementKey,
		Name:               "result",
		Value:              resultJson,
	}))

	// Complete element (should apply output mappings)
	intents := executeCompletion(t, store, &intent.CompleteElementIntent{
		Header: intent.Header{
			Key:                301,
			Origin:             intent.Internal,
			ProcessInstanceKey: piKey,
		},
		ElementInstanceKey: elementKey,
	})
	require.NotEmpty(t, intents) // Should have ActivateElement for next node

	// Check that output mapping created variable in parent scope
	v, err := store.Variables().FindByName(context.Background(), piKey, "taskResult")
	require.NoError(t, err)
	require.NotNil(t, v, "output mapping should create taskResult variable")

	var val string
	require.NoError(t, json.Unmarshal(v.Value, &val))
	assert.Equal(t, "completed successfully", val)
}

func TestSequenceFlow_ConditionExpression(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	piKey := uint64(100)
	seedProcessInstance(t, store, piKey, storage.ProcessInstanceActive)

	// Set scope variable
	amountJson, _ := json.Marshal(1500)
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: piKey,
		ScopeKey:           piKey,
		Name:               "amount",
		Value:              amountJson,
	}))

	// Test: condition is true → should emit ActivateElement
	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(takeSequenceFlow),
		&intent.TakeSequenceFlowIntent{
			Header:               intent.Header{Key: 10, ProcessInstanceKey: piKey},
			ProcessDefinitionKey: 1,
			TargetElementId:      "approved",
			ConditionExpression:  "= amount > 1000",
			FlowScopeKey:         piKey,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)
	activate, ok := intents[0].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "approved", activate.ElementId)

	// Test: condition is false → no intents
	intents, err = store.Execute(context.Background(), AsHandler(
		Typed(takeSequenceFlow),
		&intent.TakeSequenceFlowIntent{
			Header:               intent.Header{Key: 11, ProcessInstanceKey: piKey},
			ProcessDefinitionKey: 1,
			TargetElementId:      "rejected",
			ConditionExpression:  "= amount < 100",
			FlowScopeKey:         piKey,
		},
	))
	require.NoError(t, err)
	assert.Empty(t, intents, "false condition should not produce intents")
}

func TestSequenceFlow_NoCondition(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	piKey := uint64(100)
	seedProcessInstance(t, store, piKey, storage.ProcessInstanceActive)

	// No condition → always take the flow
	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(takeSequenceFlow),
		&intent.TakeSequenceFlowIntent{
			Header:               intent.Header{Key: 10, ProcessInstanceKey: piKey},
			ProcessDefinitionKey: 1,
			TargetElementId:      "next",
			FlowScopeKey:         piKey,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)
}
