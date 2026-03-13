package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_LinearProcess(t *testing.T) {
	// start -> serviceTask -> end
	bmi := CreateExecutableProcess("linear").
		StartEvent("start").
		ServiceTask("task1").Name("Do Work").
		EndEvent("end").
		Done()

	require.NotNil(t, bmi)

	// Verify process
	proc, ok := GetTypedElementById[Process](bmi.ModelInstance, "linear")
	require.True(t, ok)
	assert.True(t, proc.IsExecutable())

	// Verify elements exist
	start, ok := GetTypedElementById[StartEvent](bmi.ModelInstance, "start")
	require.True(t, ok)
	assert.NotNil(t, start)

	task, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "task1")
	require.True(t, ok)
	assert.Equal(t, "Do Work", task.GetName())

	end, ok := GetTypedElementById[EndEvent](bmi.ModelInstance, "end")
	require.True(t, ok)
	assert.NotNil(t, end)

	// Verify sequence flows
	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	assert.Len(t, flows, 2) // start->task, task->end

	// Verify outgoing/incoming
	assert.Len(t, start.GetOutgoingSequenceFlows(), 1)
	assert.Len(t, task.GetIncomingSequenceFlows(), 1)
	assert.Len(t, task.GetOutgoingSequenceFlows(), 1)
	assert.Len(t, end.GetIncomingSequenceFlows(), 1)

	// Verify flow connections
	startFlow := start.GetOutgoingSequenceFlows()[0]
	assert.Equal(t, "start", startFlow.GetSourceRef())
	assert.Equal(t, "task1", startFlow.GetTargetRef())
}

func TestBuilder_ExclusiveGateway(t *testing.T) {
	bmi := CreateExecutableProcess("gwProcess").
		StartEvent("start").
		ExclusiveGateway("decision").
		ConditionExpression("=isValid").
		ServiceTask("process").Name("Process Order").
		EndEvent("done").
		MoveToLastGateway().
		DefaultFlow().
		EndEvent("rejected").
		Done()

	require.NotNil(t, bmi)

	// Verify gateway
	gw, ok := GetTypedElementById[ExclusiveGateway](bmi.ModelInstance, "decision")
	require.True(t, ok)
	assert.NotNil(t, gw)

	// Gateway should have 2 outgoing flows
	outgoing := gw.GetOutgoingSequenceFlows()
	assert.Len(t, outgoing, 2)

	// Find conditional flow (to "process") and default flow (to "rejected")
	var condFlow, defaultFlow SequenceFlow
	for _, f := range outgoing {
		if f.GetTargetRef() == "process" {
			condFlow = f
		}
		if f.GetTargetRef() == "rejected" {
			defaultFlow = f
		}
	}

	require.NotNil(t, condFlow)
	cond := condFlow.GetConditionExpression()
	require.NotNil(t, cond)
	assert.Equal(t, "=isValid", cond.GetTextContent())

	require.NotNil(t, defaultFlow)
	// Check default flow is set on gateway
	gwDefault := gw.GetDefaultFlow()
	require.NotNil(t, gwDefault)
	assert.Equal(t, defaultFlow.GetId(), gwDefault.GetId())

	// Verify end events
	endEvents := GetTypedElements[EndEvent](bmi.ModelInstance)
	assert.Len(t, endEvents, 2)
}

func TestBuilder_ParallelGateway(t *testing.T) {
	bmi := CreateExecutableProcess("parallel").
		StartEvent("start").
		ParallelGateway("fork").
		ServiceTask("taskA").Name("Task A").
		ParallelGateway("join").
		EndEvent("end").
		MoveToNode("fork").
		ServiceTask("taskB").Name("Task B").
		ConnectTo("join").
		Done()

	require.NotNil(t, bmi)

	// Verify fork gateway has 2 outgoing
	fork, ok := GetTypedElementById[ParallelGateway](bmi.ModelInstance, "fork")
	require.True(t, ok)
	assert.Len(t, fork.GetOutgoingSequenceFlows(), 2)

	// Verify join gateway has 2 incoming
	join, ok := GetTypedElementById[ParallelGateway](bmi.ModelInstance, "join")
	require.True(t, ok)
	assert.Len(t, join.GetIncomingSequenceFlows(), 2)

	// Verify tasks
	taskA, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "taskA")
	require.True(t, ok)
	assert.Equal(t, "Task A", taskA.GetName())

	taskB, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "taskB")
	require.True(t, ok)
	assert.Equal(t, "Task B", taskB.GetName())
}

func TestBuilder_RoundTrip(t *testing.T) {
	// Build -> serialize -> parse -> verify
	bmi := CreateExecutableProcess("roundtrip").
		StartEvent("s").
		UserTask("ut").Name("Review").
		EndEvent("e").
		Done()

	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	// Re-parse
	bmi2, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	// Verify structure
	proc, ok := GetTypedElementById[Process](bmi2.ModelInstance, "roundtrip")
	require.True(t, ok)
	assert.True(t, proc.IsExecutable())

	ut, ok := GetTypedElementById[UserTask](bmi2.ModelInstance, "ut")
	require.True(t, ok)
	assert.Equal(t, "Review", ut.GetName())

	flows := GetTypedElements[SequenceFlow](bmi2.ModelInstance)
	assert.Len(t, flows, 2)
}

func TestBuilder_XMLOutput(t *testing.T) {
	bmi := CreateExecutableProcess("xmltest").
		StartEvent("start").
		ServiceTask("svc").Name("My Service").
		EndEvent("end").
		Done()

	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	assert.Contains(t, xmlStr, `id="xmltest"`)
	assert.Contains(t, xmlStr, "startEvent")
	assert.Contains(t, xmlStr, "serviceTask")
	assert.Contains(t, xmlStr, "endEvent")
	assert.Contains(t, xmlStr, "sequenceFlow")
	assert.Contains(t, xmlStr, `name="My Service"`)
	assert.Contains(t, xmlStr, "sourceRef")
	assert.Contains(t, xmlStr, "targetRef")
}

func TestBuilder_MoveToNode(t *testing.T) {
	bmi := CreateExecutableProcess("moveTest").
		StartEvent("start").
		ServiceTask("task1").
		EndEvent("end1").
		MoveToNode("task1").
		EndEvent("end2").
		Done()

	// task1 should have 2 outgoing flows
	task, ok := GetTypedElementById[ServiceTask](bmi.ModelInstance, "task1")
	require.True(t, ok)
	assert.Len(t, task.GetOutgoingSequenceFlows(), 2)
}

func TestBuilder_ConnectTo(t *testing.T) {
	bmi := CreateExecutableProcess("connectTest").
		StartEvent("start").
		ServiceTask("task1").
		EndEvent("end").
		MoveToNode("start").
		ConnectTo("end").
		Done()

	// start should have 2 outgoing (to task1 and to end)
	start, ok := GetTypedElementById[StartEvent](bmi.ModelInstance, "start")
	require.True(t, ok)
	assert.Len(t, start.GetOutgoingSequenceFlows(), 2)

	// end should have 2 incoming
	end, ok := GetTypedElementById[EndEvent](bmi.ModelInstance, "end")
	require.True(t, ok)
	assert.Len(t, end.GetIncomingSequenceFlows(), 2)
}

func TestBuilder_UserTask(t *testing.T) {
	bmi := CreateExecutableProcess("userTaskProc").
		StartEvent("start").
		UserTask("review").Name("Manual Review").Implementation("##unspecified").
		EndEvent("end").
		Done()

	ut, ok := GetTypedElementById[UserTask](bmi.ModelInstance, "review")
	require.True(t, ok)
	assert.Equal(t, "Manual Review", ut.GetName())
	assert.Equal(t, "##unspecified", ut.GetImplementation())
}

func TestBuilder_ScriptTask(t *testing.T) {
	bmi := CreateExecutableProcess("scriptProc").
		StartEvent("start").
		ScriptTask("script").Name("Run Script").ScriptFormat("groovy").
		EndEvent("end").
		Done()

	st, ok := GetTypedElementById[ScriptTask](bmi.ModelInstance, "script")
	require.True(t, ok)
	assert.Equal(t, "Run Script", st.GetName())
	assert.Equal(t, "groovy", st.GetScriptFormat())
}

func TestBuilder_CallActivity(t *testing.T) {
	bmi := CreateExecutableProcess("callerProc").
		StartEvent("start").
		CallActivity("call").Name("Call Sub").CalledElement("subProcess").
		EndEvent("end").
		Done()

	ca, ok := GetTypedElementById[CallActivity](bmi.ModelInstance, "call")
	require.True(t, ok)
	assert.Equal(t, "Call Sub", ca.GetName())
	assert.Equal(t, "subProcess", ca.GetCalledElement())
}

func TestBuilder_AutoGenerateId(t *testing.T) {
	bmi := CreateExecutableProcess("autoId").
		StartEvent("").
		EndEvent("").
		Done()

	startEvents := GetTypedElements[StartEvent](bmi.ModelInstance)
	require.Len(t, startEvents, 1)
	assert.NotEmpty(t, startEvents[0].GetId())

	endEvents := GetTypedElements[EndEvent](bmi.ModelInstance)
	require.Len(t, endEvents, 1)
	assert.NotEmpty(t, endEvents[0].GetId())

	// IDs should be different
	assert.NotEqual(t, startEvents[0].GetId(), endEvents[0].GetId())
}
