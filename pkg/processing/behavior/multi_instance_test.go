package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

const bpmnWithParallelMI = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:serviceTask id="task1" name="Process Item">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="item-worker"/>
        <zeebe:loopCharacteristics inputCollection="= items" inputElement="item"/>
      </bpmn:extensionElements>
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:multiInstanceLoopCharacteristics/>
    </bpmn:serviceTask>
    <bpmn:endEvent id="end">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="task1"/>
    <bpmn:sequenceFlow id="flow2" sourceRef="task1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

const bpmnWithSequentialMI = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="test-process" isExecutable="true">
    <bpmn:serviceTask id="task1" name="Process Item">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="item-worker"/>
        <zeebe:loopCharacteristics inputCollection="= items" inputElement="item"/>
      </bpmn:extensionElements>
      <bpmn:multiInstanceLoopCharacteristics isSequential="true"/>
    </bpmn:serviceTask>
  </bpmn:process>
</bpmn:definitions>`

func TestMultiInstance_Parallel_CreatesChildInstances(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedBpmnFromXML(t, store, 1, bpmnWithParallelMI)

	// Set items variable
	itemsJSON, _ := json.Marshal([]string{"a", "b", "c"})
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "items",
		Value:              itemsJSON,
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(activateElement),
		&intent.ActivateElementIntent{
			Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
			ProcessDefinitionKey: 1,
			ElementId:            "task1",
			ElementType:          "serviceTask",
			FlowScopeKey:         10,
			JobType:              "item-worker",
		},
	))
	require.NoError(t, err)

	// Should have: 1 SetVariables (MI metadata) + 3 ActivateElement (children)
	require.Len(t, intents, 4)

	// First intent is SetVariables for MI metadata
	_, ok := intents[0].(*intent.SetVariablesIntent)
	require.True(t, ok)

	// Next 3 are ActivateElement for each child
	for idx := 1; idx <= 3; idx++ {
		child, ok := intents[idx].(*intent.ActivateElementIntent)
		require.True(t, ok)
		assert.Equal(t, "task1", child.ElementId)
		assert.Equal(t, "serviceTask", child.ElementType)
		assert.Equal(t, "item", child.MIInputVariable)
		assert.Equal(t, uint64(100), child.FlowScopeKey) // MI body is scope
	}

	// Check that each child has different item values
	values := make([]string, 3)
	for idx := 1; idx <= 3; idx++ {
		child := intents[idx].(*intent.ActivateElementIntent)
		var v string
		require.NoError(t, json.Unmarshal(child.MIInputValue, &v))
		values[idx-1] = v
	}
	assert.Equal(t, []string{"a", "b", "c"}, values)
}

func TestMultiInstance_Sequential_CreatesOneChild(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedBpmnFromXML(t, store, 1, bpmnWithSequentialMI)

	itemsJSON, _ := json.Marshal([]int{1, 2, 3})
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "items",
		Value:              itemsJSON,
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(activateElement),
		&intent.ActivateElementIntent{
			Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
			ProcessDefinitionKey: 1,
			ElementId:            "task1",
			ElementType:          "serviceTask",
			FlowScopeKey:         10,
			JobType:              "item-worker",
		},
	))
	require.NoError(t, err)

	// Should have: 1 SetVariables + 1 ActivateElement (only first child)
	require.Len(t, intents, 2)

	child, ok := intents[1].(*intent.ActivateElementIntent)
	require.True(t, ok)
	assert.Equal(t, "item", child.MIInputVariable)
	assert.Equal(t, 0, child.MIIndex)
}

func TestMultiInstance_EmptyCollection_CompletesImmediately(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))
	seedBpmnFromXML(t, store, 1, bpmnWithParallelMI)

	itemsJSON, _ := json.Marshal([]string{})
	require.NoError(t, store.Variables().Create(context.Background(), &storage.Variable{
		Key:                1,
		ProcessInstanceKey: 10,
		ScopeKey:           10,
		Name:               "items",
		Value:              itemsJSON,
	}))

	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(activateElement),
		&intent.ActivateElementIntent{
			Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
			ProcessDefinitionKey: 1,
			ElementId:            "task1",
			ElementType:          "serviceTask",
			FlowScopeKey:         10,
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)

	_, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
}

func TestGetMultiInstanceInfo(t *testing.T) {
	bmi, err := bpmn_model.ReadFromString(bpmnWithParallelMI)
	require.NoError(t, err)

	mi := getMultiInstanceInfo(bmi, "task1")
	require.NotNil(t, mi)
	assert.Equal(t, "= items", mi.InputCollection)
	assert.Equal(t, "item", mi.InputElement)
	assert.False(t, mi.IsSequential)
}

func TestGetMultiInstanceInfo_Sequential(t *testing.T) {
	bmi, err := bpmn_model.ReadFromString(bpmnWithSequentialMI)
	require.NoError(t, err)

	mi := getMultiInstanceInfo(bmi, "task1")
	require.NotNil(t, mi)
	assert.True(t, mi.IsSequential)
}

func TestGetMultiInstanceInfo_NoMI(t *testing.T) {
	bmi := bpmn_model.CreateExecutableProcess("test").
		StartEvent("start").
		ServiceTask("task1").ZeebeJobType("worker").
		EndEvent("end").
		Done()

	mi := getMultiInstanceInfo(bmi, "task1")
	assert.Nil(t, mi)
}
