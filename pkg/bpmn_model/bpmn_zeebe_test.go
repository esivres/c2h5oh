package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Builder tests ---

func TestZeebe_ServiceTaskBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("zeebeProc").
		StartEvent("start").
		ServiceTask("task1").
		Name("My Task").
		ZeebeJobType("my-worker").
		ZeebeJobRetries("5").
		ZeebeTaskHeader("headerKey", "headerValue").
		ZeebeTaskHeader("anotherKey", "anotherValue").
		ZeebeInput("=sourceVar", "inputVar").
		ZeebeOutput("=result", "outputVar").
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "task1")
	require.True(t, ok)

	// ZeebeTaskDefinition
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "my-worker", td.GetType())
	assert.Equal(t, "5", td.GetRetries())

	// ZeebeTaskHeaders
	th, ok := GetSingleExtensionElement[ZeebeTaskHeaders](task)
	require.True(t, ok)
	headers := th.GetHeaders()
	require.Len(t, headers, 2)
	assert.Equal(t, "headerKey", headers[0].GetKey())
	assert.Equal(t, "headerValue", headers[0].GetValue())
	assert.Equal(t, "anotherKey", headers[1].GetKey())

	// ZeebeIoMapping
	ioMap, ok := GetSingleExtensionElement[ZeebeIoMapping](task)
	require.True(t, ok)
	inputs := ioMap.GetInputs()
	require.Len(t, inputs, 1)
	assert.Equal(t, "=sourceVar", inputs[0].GetSource())
	assert.Equal(t, "inputVar", inputs[0].GetTarget())
	outputs := ioMap.GetOutputs()
	require.Len(t, outputs, 1)
	assert.Equal(t, "=result", outputs[0].GetSource())
	assert.Equal(t, "outputVar", outputs[0].GetTarget())
}

func TestZeebe_UserTaskBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("userProc").
		StartEvent("start").
		UserTask("review").
		Name("Review Task").
		ZeebeAssignee("john").
		ZeebeCandidateGroups("managers,admins").
		ZeebeCandidateUsers("alice,bob").
		ZeebeFormId("myForm").
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[UserTask](bmi.ModelInstance, "review")
	require.True(t, ok)

	ad, ok := GetSingleExtensionElement[ZeebeAssignmentDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "john", ad.GetAssignee())
	assert.Equal(t, "managers,admins", ad.GetCandidateGroups())
	assert.Equal(t, "alice,bob", ad.GetCandidateUsers())

	fd, ok := GetSingleExtensionElement[ZeebeFormDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "myForm", fd.GetFormId())
}

func TestZeebe_CallActivityBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("callerProc").
		StartEvent("start").
		CallActivity("call").
		ZeebeProcessId("childProcess").
		ZeebePropagateAllChildVariables(false).
		EndEvent("end").
		Done()

	ca, ok := GetTypedElementById[CallActivity](bmi.ModelInstance, "call")
	require.True(t, ok)

	ce, ok := GetSingleExtensionElement[ZeebeCalledElement](ca)
	require.True(t, ok)
	assert.Equal(t, "childProcess", ce.GetProcessId())
	assert.False(t, ce.GetPropagateAllChildVariables())
}

func TestZeebe_BusinessRuleTaskBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("dmnProc").
		StartEvent("start").
		BusinessRuleTask("dmn").
		ZeebeCalledDecision("myDecision", "result").
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[BusinessRuleTask](bmi.ModelInstance, "dmn")
	require.True(t, ok)

	cd, ok := GetSingleExtensionElement[ZeebeCalledDecision](task)
	require.True(t, ok)
	assert.Equal(t, "myDecision", cd.GetDecisionId())
	assert.Equal(t, "result", cd.GetResultVariable())
}

func TestZeebe_ScriptTaskBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("scriptProc").
		StartEvent("start").
		ScriptTask("script").
		ZeebeScript("=myExpression", "result").
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[ScriptTask](bmi.ModelInstance, "script")
	require.True(t, ok)

	s, ok := GetSingleExtensionElement[ZeebeScript](task)
	require.True(t, ok)
	assert.Equal(t, "=myExpression", s.GetExpression())
	assert.Equal(t, "result", s.GetResultVariable())
}

func TestZeebe_ReceiveTaskBuilder(t *testing.T) {
	bmi := CreateExecutableProcess("msgProc").
		StartEvent("start").
		ReceiveTask("recv").
		ZeebeCorrelationKey("=orderId").
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[ReceiveTask](bmi.ModelInstance, "recv")
	require.True(t, ok)

	sub, ok := GetSingleExtensionElement[ZeebeSubscription](task)
	require.True(t, ok)
	assert.Equal(t, "=orderId", sub.GetCorrelationKey())
}

// --- Parsing tests ---

const zeebeXML = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:zeebe="http://camunda.org/schema/zeebe/1.0"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:serviceTask id="task1" name="My Task">
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="my-type" retries="3"/>
        <zeebe:ioMapping>
          <zeebe:input source="=a" target="b"/>
          <zeebe:output source="=c" target="d"/>
        </zeebe:ioMapping>
        <zeebe:taskHeaders>
          <zeebe:header key="k1" value="v1"/>
          <zeebe:header key="k2" value="v2"/>
        </zeebe:taskHeaders>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:userTask id="user1">
      <bpmn:extensionElements>
        <zeebe:assignmentDefinition assignee="demo" candidateGroups="team-a"/>
        <zeebe:formDefinition formId="form1"/>
      </bpmn:extensionElements>
    </bpmn:userTask>
    <bpmn:callActivity id="call1">
      <bpmn:extensionElements>
        <zeebe:calledElement processId="child" propagateAllChildVariables="false"/>
      </bpmn:extensionElements>
    </bpmn:callActivity>
  </bpmn:process>
</bpmn:definitions>`

func TestZeebe_ParseServiceTask(t *testing.T) {
	bmi, err := ReadFromString(zeebeXML)
	require.NoError(t, err)

	task, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "task1")
	require.True(t, ok)

	// TaskDefinition
	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "my-type", td.GetType())
	assert.Equal(t, "3", td.GetRetries())

	// IoMapping
	ioMap, ok := GetSingleExtensionElement[ZeebeIoMapping](task)
	require.True(t, ok)
	inputs := ioMap.GetInputs()
	require.Len(t, inputs, 1)
	assert.Equal(t, "=a", inputs[0].GetSource())
	assert.Equal(t, "b", inputs[0].GetTarget())
	outputs := ioMap.GetOutputs()
	require.Len(t, outputs, 1)
	assert.Equal(t, "=c", outputs[0].GetSource())
	assert.Equal(t, "d", outputs[0].GetTarget())

	// TaskHeaders
	th, ok := GetSingleExtensionElement[ZeebeTaskHeaders](task)
	require.True(t, ok)
	headers := th.GetHeaders()
	require.Len(t, headers, 2)
	assert.Equal(t, "k1", headers[0].GetKey())
	assert.Equal(t, "v1", headers[0].GetValue())
	assert.Equal(t, "k2", headers[1].GetKey())
	assert.Equal(t, "v2", headers[1].GetValue())
}

func TestZeebe_ParseUserTask(t *testing.T) {
	bmi, err := ReadFromString(zeebeXML)
	require.NoError(t, err)

	task, ok := GetTypedElementById[UserTask](bmi.ModelInstance, "user1")
	require.True(t, ok)

	ad, ok := GetSingleExtensionElement[ZeebeAssignmentDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "demo", ad.GetAssignee())
	assert.Equal(t, "team-a", ad.GetCandidateGroups())

	fd, ok := GetSingleExtensionElement[ZeebeFormDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "form1", fd.GetFormId())
}

func TestZeebe_ParseCallActivity(t *testing.T) {
	bmi, err := ReadFromString(zeebeXML)
	require.NoError(t, err)

	ca, ok := GetTypedElementById[CallActivity](bmi.ModelInstance, "call1")
	require.True(t, ok)

	ce, ok := GetSingleExtensionElement[ZeebeCalledElement](ca)
	require.True(t, ok)
	assert.Equal(t, "child", ce.GetProcessId())
	assert.False(t, ce.GetPropagateAllChildVariables())
}

// --- Round-trip test ---

func TestZeebe_RoundTrip(t *testing.T) {
	// Build with Zeebe extensions
	bmi := CreateExecutableProcess("roundtrip").
		StartEvent("start").
		ServiceTask("svc").
		ZeebeJobType("worker").
		ZeebeTaskHeader("key", "val").
		EndEvent("end").
		Done()

	// Serialize
	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	assert.Contains(t, xmlStr, "zeebe:taskDefinition")
	assert.Contains(t, xmlStr, `type="worker"`)
	assert.Contains(t, xmlStr, "zeebe:header")
	assert.Contains(t, xmlStr, `key="key"`)

	// Re-parse
	bmi2, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	task, ok := GetTypedElementById[ServiceTask](bmi2.ModelInstance, "svc")
	require.True(t, ok)

	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "worker", td.GetType())

	th, ok := GetSingleExtensionElement[ZeebeTaskHeaders](task)
	require.True(t, ok)
	headers := th.GetHeaders()
	require.Len(t, headers, 1)
	assert.Equal(t, "key", headers[0].GetKey())
	assert.Equal(t, "val", headers[0].GetValue())
}

func TestZeebe_ExtensionElements_GetOrCreate(t *testing.T) {
	bmi := CreateExecutableProcess("extTest").Done()
	proc, _ := GetTypedElementById[Process](bmi.ModelInstance, "extTest")

	// Create a service task manually
	taskType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_SERVICE_TASK)
	inst, _ := bmi.NewInstance(taskType)
	task := inst.(ServiceTask)
	task.SetId("manualTask")
	proc.AddFlowElement(task)

	// First call creates extension elements
	ee1 := task.GetOrCreateExtensionElements()
	require.NotNil(t, ee1)

	// Second call returns the same
	ee2 := task.GetOrCreateExtensionElements()
	require.NotNil(t, ee2)
	assert.Equal(t, ee1.GetDomElement().Unwrap(), ee2.GetDomElement().Unwrap())
}

func TestZeebe_MultipleJobTypeCallsReuseSameElement(t *testing.T) {
	bmi := CreateExecutableProcess("reuseTest").
		StartEvent("start").
		ServiceTask("svc").
		ZeebeJobType("first").
		ZeebeJobType("second"). // Should overwrite, not create a new element
		EndEvent("end").
		Done()

	task, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "svc")
	require.True(t, ok)

	td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](task)
	require.True(t, ok)
	assert.Equal(t, "second", td.GetType()) // last value wins
}
