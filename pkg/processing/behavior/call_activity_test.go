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

func TestCallActivity_CreatesChildProcessInstance(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Deploy parent process with call activity
	parentBmi := bpmn_model.CreateExecutableProcess("parent-process").
		StartEvent("start").
		CallActivity("call1").ZeebeProcessId("child-process").
		EndEvent("end").
		Done()
	parentContent, err := parentBmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           1,
		BpmnProcessId: "parent-process",
		Version:       1,
		Content:       parentContent,
		ContentHash:   []byte("parent-hash"),
		DeployedAt:    time.Now(),
	}))

	// Deploy child process
	childBmi := bpmn_model.CreateExecutableProcess("child-process").
		StartEvent("childStart").
		ServiceTask("childTask").ZeebeJobType("child-worker").
		EndEvent("childEnd").
		Done()
	childContent, err := childBmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           2,
		BpmnProcessId: "child-process",
		Version:       1,
		Content:       childContent,
		ContentHash:   []byte("child-hash"),
		DeployedAt:    time.Now(),
	}))

	// Activate call activity
	intents := executeActivation(t, store, &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "call1",
		ElementType:          "callActivity",
		FlowScopeKey:         10,
	})
	require.Len(t, intents, 1)

	create, ok := intents[0].(*intent.CreateProcessInstanceIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(2), create.ProcessDefinitionKey)
	assert.Equal(t, "child-process", create.BpmnProcessId)
	assert.Equal(t, uint64(10), create.ParentKey)
	assert.Equal(t, uint64(100), create.ParentElementKey)
}

func TestCallActivity_ChildComplete_CompletesParentElement(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Create a child process instance linked to parent
	require.NoError(t, store.ProcessInstances().CreateInstance(context.Background(), &storage.ProcessInstance{
		Key:                  20,
		ProcessDefinitionKey: 2,
		BpmnProcessId:        "child-process",
		ParentKey:            10,
		ParentElementKey:     100, // call activity element instance key
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}))

	// Complete child process
	intents, err := store.Execute(context.Background(), AsHandler(
		Typed(completeProcessInstance),
		&intent.CompleteProcessInstanceIntent{
			Header: intent.Header{Key: 50, ProcessInstanceKey: 20},
		},
	))
	require.NoError(t, err)
	require.Len(t, intents, 1)

	// Should complete the parent call activity element
	complete, ok := intents[0].(*intent.CompleteElementIntent)
	require.True(t, ok)
	assert.Equal(t, uint64(100), complete.ElementInstanceKey)
	assert.Equal(t, uint64(10), complete.ProcessInstanceKey)
}

func TestCallActivity_CalledProcessNotFound(t *testing.T) {
	store := sqlstore.NewStore(openTestDB(t))

	// Deploy parent process with call activity pointing to non-existent process
	parentBmi := bpmn_model.CreateExecutableProcess("parent-process").
		StartEvent("start").
		CallActivity("call1").ZeebeProcessId("non-existent-process").
		EndEvent("end").
		Done()
	parentContent, err := parentBmi.GetDocument().WriteToBytes()
	require.NoError(t, err)
	require.NoError(t, store.ProcessDefinitions().Create(context.Background(), &storage.ProcessDefinition{
		Key:           1,
		BpmnProcessId: "parent-process",
		Version:       1,
		Content:       parentContent,
		ContentHash:   []byte("parent-hash"),
		DeployedAt:    time.Now(),
	}))

	ctx := context.Background()
	i := &intent.ActivateElementIntent{
		Header:               intent.Header{Key: 100, ProcessInstanceKey: 10},
		ProcessDefinitionKey: 1,
		ElementId:            "call1",
		ElementType:          "callActivity",
		FlowScopeKey:         10,
	}

	// Phase 1
	phase1, err := store.Execute(ctx, AsHandler(Typed(activateElement), i))
	require.NoError(t, err)
	activated := phase1[0].(*intent.ElementActivatedIntent)

	// Phase 2 — expect error
	reg := DefaultRegistry()
	b := reg.LookupWithElementType(intent.ElementActivated, activated.ElementType)
	_, err = store.Execute(ctx, AsHandler(b, activated))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
