package bpmn_asserts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- JobAssert ---

func TestJobAssert_HasType(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "kamundarf:adhoc:v1", "task1", 100, 3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForJob(t, stream, 10).WithContext(ctx).
		HasType("kamundarf:adhoc:v1").
		HasElementID("task1").
		HasBpmnProcessID("testProcess").
		HasRetries(3)
}

func TestJobAssert_HasType_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "kamundarf:adhoc:v1", "task1", 100, 3),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	catchFatal(func() {
		ForJob(fc, stream, 10).WithContext(ctx).HasType("wrong:type")
	})
	assert.True(t, fc.fataled)
}

func TestJobAssert_HasDeadline(t *testing.T) {
	stream := NewRecordStream()
	rec := makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3)
	// Set a specific deadline in the job value
	rec = makeRecord("JOB", "EVENT", "CREATED", 10, JobValue{
		Type:               "myJob",
		ElementID:          "task1",
		ProcessInstanceKey: 100,
		Retries:            3,
		BpmnProcessID:      "testProcess",
		Deadline:           1000000,
	})
	feedRecords(stream, rec)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForJob(t, stream, 10).WithContext(ctx).HasDeadline(1000000, 100)
}

func TestJobAssert_HasNoIncidents(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForJob(t, stream, 10).WithContext(ctx).HasNoIncidents()
}

func TestJobAssert_HasAnyIncidents(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "no retries", "task1", 10),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForJob(t, stream, 10).WithContext(ctx).HasAnyIncidents()
}

func TestJobAssert_ExtractingVariables(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	vars := ForJob(t, stream, 10).WithContext(ctx).ExtractingVariables()
	assert.Equal(t, "value", vars["input"])
}

func TestJobAssert_ExtractingHeaders(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeJobRecord("CREATED", 10, "myJob", "task1", 100, 3),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	headers := ForJob(t, stream, 10).WithContext(ctx).ExtractingHeaders()
	assert.Equal(t, "val1", headers["header1"])
}

// --- IncidentAssert ---

func TestIncidentAssert_Properties(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "no retries left", "task1", 10),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForIncident(t, stream, 20).WithContext(ctx).
		HasErrorType("JOB_NO_RETRIES").
		HasErrorMessage("no retries left").
		WasRaisedInProcessInstance(100).
		OccurredOnElement("task1").
		OccurredDuringJob(10)
}

func TestIncidentAssert_IsResolved(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
		makeIncidentRecord("RESOLVED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForIncident(t, stream, 20).WithContext(ctx).IsResolved()
}

func TestIncidentAssert_IsUnresolved(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
	)

	ForIncident(t, stream, 20).IsUnresolved()
}

func TestIncidentAssert_IsUnresolved_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeIncidentRecord("CREATED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
		makeIncidentRecord("RESOLVED", 20, 100, "JOB_NO_RETRIES", "msg", "task1", 10),
	)

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForIncident(fc, stream, 20).IsUnresolved()
	})
	assert.True(t, fc.fataled)
}

// --- DeploymentAssert ---

func TestDeploymentAssert_ContainsProcesses(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeDeploymentRecord("CREATED", 1, []string{"processA", "processB"}, []string{"a.bpmn", "b.bpmn"}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForDeployment(t, stream, 1).WithContext(ctx).
		ContainsProcessesByBpmnProcessID("processA", "processB").
		ContainsProcessesByResourceName("a.bpmn", "b.bpmn")
}

func TestDeploymentAssert_ContainsProcesses_Missing(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeDeploymentRecord("CREATED", 1, []string{"processA"}, []string{"a.bpmn"}),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	catchFatal(func() {
		ForDeployment(fc, stream, 1).WithContext(ctx).ContainsProcessesByBpmnProcessID("processB")
	})
	assert.True(t, fc.fataled)
}

func TestDeploymentAssert_ExtractingProcess(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeDeploymentRecord("CREATED", 1, []string{"processA"}, []string{"a.bpmn"}),
		makeProcessDefinitionRecord("CREATED", 100, "processA", 1, "a.bpmn"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForDeployment(t, stream, 1).WithContext(ctx).
		ExtractingProcessByBpmnProcessID("processA").
		HasBpmnProcessID("processA").
		HasVersion(1)
}

// --- MessageAssert ---

func TestMessageAssert_HasBeenCorrelated(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageCorrelatedSubscription(1, 100, 500, "paymentReceived"),
	)

	ForMessage(t, stream, 500).HasBeenCorrelated()
}

func TestMessageAssert_HasNotBeenCorrelated(t *testing.T) {
	stream := NewRecordStream()

	ForMessage(t, stream, 500).HasNotBeenCorrelated()
}

func TestMessageAssert_HasCreatedProcessInstance(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageStartEventCorrelated(1, 300, 500, "startMsg"),
	)

	ForMessage(t, stream, 500).HasCreatedProcessInstance()
}

func TestMessageAssert_HasNotCreatedProcessInstance(t *testing.T) {
	stream := NewRecordStream()

	ForMessage(t, stream, 500).HasNotCreatedProcessInstance()
}

func TestMessageAssert_HasExpired(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageRecord("EXPIRED", 500, "myMsg", "corr-123"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForMessage(t, stream, 500).WithContext(ctx).HasExpired()
}

func TestMessageAssert_HasNotExpired(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageRecord("PUBLISHED", 500, "myMsg", "corr-123"),
	)

	ForMessage(t, stream, 500).HasNotExpired()
}

func TestMessageAssert_ExtractingProcessInstance(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeMessageCorrelatedSubscription(1, 100, 500, "paymentReceived"),
		makeProcessInstanceRecord("ELEMENT_COMPLETED", 2, 100, "testProcess", "PROCESS"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForMessage(t, stream, 500).
		ExtractingProcessInstance().
		WithContext(ctx).
		IsCompleted()
}

// --- ProcessDefinitionAssert ---

func TestProcessDefinitionAssert_Properties(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessDefinitionRecord("CREATED", 100, "myProcess", 1, "process.bpmn"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForProcessDefinition(t, stream, 100).WithContext(ctx).
		HasBpmnProcessID("myProcess").
		HasVersion(1).
		HasResourceName("process.bpmn")
}

func TestProcessDefinitionAssert_HasInstances(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessDefinitionRecord("CREATED", 100, "myProcess", 1, "process.bpmn"),
		makeProcessRecord("ELEMENT_ACTIVATED", 1, 200, "myProcess", "PROCESS", "myProcess"),
		makeProcessRecord("ELEMENT_ACTIVATED", 2, 300, "myProcess", "PROCESS", "myProcess"),
	)

	// Need to set processDefinitionKey on the process instance records
	// The helper doesn't set it, so let me create proper records
	stream2 := NewRecordStream()
	feedRecords(stream2,
		makeProcessDefinitionRecord("CREATED", 100, "myProcess", 1, "process.bpmn"),
	)

	piRec1 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 1, ProcessInstanceValue{
		ProcessInstanceKey:   200,
		ProcessDefinitionKey: 100,
		BpmnElementType:      "PROCESS",
		BpmnProcessID:        "myProcess",
		ElementID:            "myProcess",
	})
	piRec2 := makeRecord("PROCESS_INSTANCE", "EVENT", "ELEMENT_ACTIVATED", 2, ProcessInstanceValue{
		ProcessInstanceKey:   300,
		ProcessDefinitionKey: 100,
		BpmnElementType:      "PROCESS",
		BpmnProcessID:        "myProcess",
		ElementID:            "myProcess",
	})
	feedRecords(stream2, piRec1, piRec2)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForProcessDefinition(t, stream2, 100).WithContext(ctx).
		HasAnyInstances().
		HasInstances(2)
}

func TestProcessDefinitionAssert_HasNoInstances(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeProcessDefinitionRecord("CREATED", 100, "myProcess", 1, "process.bpmn"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForProcessDefinition(t, stream, 100).WithContext(ctx).HasNoInstances()
}

// --- FormAssert ---

func TestFormAssert_Properties(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeFormRecord("CREATED", 50, "loginForm", 1, "login.form"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ForForm(t, stream, 50).WithContext(ctx).
		HasFormID("loginForm").
		HasFormKey(50).
		HasVersion(1).
		HasResourceName("login.form")
}

func TestFormAssert_HasFormID_Fails(t *testing.T) {
	stream := NewRecordStream()
	feedRecords(stream,
		makeFormRecord("CREATED", 50, "loginForm", 1, "login.form"),
	)

	fc := &fatalCatcher{TB: t}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	catchFatal(func() {
		ForForm(fc, stream, 50).WithContext(ctx).HasFormID("wrongForm")
	})
	assert.True(t, fc.fataled)
}

// --- VariablesMapAssert ---

func TestVariablesMapAssert_ContainsVariable(t *testing.T) {
	vars := map[string]string{"a": `"hello"`, "b": "42"}

	ForVariables(t, vars).
		ContainsVariable("a").
		ContainsVariable("b").
		HasSize(2)
}

func TestVariablesMapAssert_DoesNotContainVariable(t *testing.T) {
	vars := map[string]string{"a": `"hello"`}

	ForVariables(t, vars).DoesNotContainVariable("b")
}

func TestVariablesMapAssert_DoesNotContainVariable_Fails(t *testing.T) {
	vars := map[string]string{"a": `"hello"`}

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForVariables(fc, vars).DoesNotContainVariable("a")
	})
	assert.True(t, fc.fataled)
}

func TestVariablesMapAssert_HasVariableWithValue(t *testing.T) {
	vars := map[string]string{"count": "42", "name": `"hello"`}

	ForVariables(t, vars).
		HasVariableWithValue("count", 42).
		HasVariableWithValue("name", "hello")
}

func TestVariablesMapAssert_IsEmpty(t *testing.T) {
	ForVariables(t, map[string]string{}).IsEmpty()
}

func TestVariablesMapAssert_IsEmpty_Fails(t *testing.T) {
	vars := map[string]string{"a": "1"}

	fc := &fatalCatcher{TB: t}
	catchFatal(func() {
		ForVariables(fc, vars).IsEmpty()
	})
	assert.True(t, fc.fataled)
}

func TestVariablesMapAssert_Raw(t *testing.T) {
	vars := map[string]string{"a": "1"}
	raw := ForVariables(t, vars).Raw()
	assert.Equal(t, vars, raw)
}
