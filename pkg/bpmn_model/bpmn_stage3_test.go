package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBpmnWithTasks = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1" name="Start"/>
    <bpmn:serviceTask id="Service_1" name="Call API" implementation="##WebService"/>
    <bpmn:userTask id="User_1" name="Review" implementation="##unspecified"/>
    <bpmn:scriptTask id="Script_1" name="Transform" scriptFormat="javascript"/>
    <bpmn:businessRuleTask id="BRule_1" name="Check Rules" implementation="##unspecified"/>
    <bpmn:sendTask id="Send_1" name="Send Email" implementation="##WebService"/>
    <bpmn:receiveTask id="Receive_1" name="Wait Message" implementation="##WebService"/>
    <bpmn:manualTask id="Manual_1" name="Manual Step"/>
    <bpmn:endEvent id="End_1" name="End"/>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="Service_1"/>
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Service_1" targetRef="End_1"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseServiceTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	tasks := GetTypedElements[ServiceTask](bmi.ModelInstance)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Service_1", tasks[0].GetId())
	assert.Equal(t, "Call API", tasks[0].GetName())
	assert.Equal(t, "##WebService", tasks[0].GetImplementation())
	assert.Equal(t, BPMN_ELEMENT_SERVICE_TASK, tasks[0].BpmnElementType())
}

func TestParseUserTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	tasks := GetTypedElements[UserTask](bmi.ModelInstance)
	require.Len(t, tasks, 1)
	assert.Equal(t, "User_1", tasks[0].GetId())
	assert.Equal(t, "Review", tasks[0].GetName())
	assert.Equal(t, "##unspecified", tasks[0].GetImplementation())
}

func TestParseScriptTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	tasks := GetTypedElements[ScriptTask](bmi.ModelInstance)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Script_1", tasks[0].GetId())
	assert.Equal(t, "javascript", tasks[0].GetScriptFormat())
}

func TestParseBusinessRuleTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	tasks := GetTypedElements[BusinessRuleTask](bmi.ModelInstance)
	require.Len(t, tasks, 1)
	assert.Equal(t, "BRule_1", tasks[0].GetId())
}

func TestParseSendReceiveTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	sends := GetTypedElements[SendTask](bmi.ModelInstance)
	require.Len(t, sends, 1)
	assert.Equal(t, "Send_1", sends[0].GetId())
	assert.Equal(t, "##WebService", sends[0].GetImplementation())

	receives := GetTypedElements[ReceiveTask](bmi.ModelInstance)
	require.Len(t, receives, 1)
	assert.Equal(t, "Receive_1", receives[0].GetId())
}

func TestParseManualTask(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	tasks := GetTypedElements[ManualTask](bmi.ModelInstance)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Manual_1", tasks[0].GetId())
	assert.Equal(t, "Manual Step", tasks[0].GetName())
}

func TestTaskTypeDistinction(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithTasks)
	require.NoError(t, err)

	// Each task type should only match its own kind
	assert.Len(t, GetTypedElements[ServiceTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[UserTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[ScriptTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[BusinessRuleTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[SendTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[ReceiveTask](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[ManualTask](bmi.ModelInstance), 1)
}

const testBpmnWithGateways = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1"/>
    <bpmn:exclusiveGateway id="ExGw_1" name="Decision" default="Flow_yes"/>
    <bpmn:parallelGateway id="ParGw_1" name="Fork"/>
    <bpmn:inclusiveGateway id="InclGw_1"/>
    <bpmn:eventBasedGateway id="EbGw_1" instantiate="true" eventGatewayType="Parallel"/>
    <bpmn:complexGateway id="CmplGw_1"/>
    <bpmn:endEvent id="End_1"/>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="ExGw_1"/>
    <bpmn:sequenceFlow id="Flow_yes" sourceRef="ExGw_1" targetRef="End_1"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseExclusiveGateway(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	gws := GetTypedElements[ExclusiveGateway](bmi.ModelInstance)
	require.Len(t, gws, 1)
	assert.Equal(t, "ExGw_1", gws[0].GetId())
	assert.Equal(t, "Decision", gws[0].GetName())

	// Test default flow ref
	defFlow := gws[0].GetDefaultFlow()
	require.NotNil(t, defFlow)
	assert.Equal(t, "Flow_yes", defFlow.GetId())
}

func TestParseParallelGateway(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	gws := GetTypedElements[ParallelGateway](bmi.ModelInstance)
	require.Len(t, gws, 1)
	assert.Equal(t, "ParGw_1", gws[0].GetId())
	assert.Equal(t, "Fork", gws[0].GetName())
}

func TestParseInclusiveGateway(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	gws := GetTypedElements[InclusiveGateway](bmi.ModelInstance)
	require.Len(t, gws, 1)
	assert.Equal(t, "InclGw_1", gws[0].GetId())
}

func TestParseEventBasedGateway(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	gws := GetTypedElements[EventBasedGateway](bmi.ModelInstance)
	require.Len(t, gws, 1)
	assert.Equal(t, "EbGw_1", gws[0].GetId())
	assert.True(t, gws[0].IsInstantiate())
	assert.Equal(t, "Parallel", gws[0].GetEventGatewayType())
}

func TestParseComplexGateway(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	gws := GetTypedElements[ComplexGateway](bmi.ModelInstance)
	require.Len(t, gws, 1)
	assert.Equal(t, "CmplGw_1", gws[0].GetId())
}

func TestGatewayTypeDistinction(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithGateways)
	require.NoError(t, err)

	assert.Len(t, GetTypedElements[ExclusiveGateway](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[ParallelGateway](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[InclusiveGateway](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[EventBasedGateway](bmi.ModelInstance), 1)
	assert.Len(t, GetTypedElements[ComplexGateway](bmi.ModelInstance), 1)
}

const testBpmnWithEvents = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:message id="Msg_1" name="OrderReceived"/>
  <bpmn:signal id="Sig_1" name="AllDone"/>
  <bpmn:error id="Err_1" name="ValidationError" errorCode="ERR_001"/>
  <bpmn:escalation id="Esc_1" name="HighPriority" escalationCode="ESC_HIGH"/>
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1">
      <bpmn:messageEventDefinition id="MsgEvtDef_1" messageRef="Msg_1"/>
    </bpmn:startEvent>
    <bpmn:serviceTask id="Task_1" name="Do Work" implementation="##WebService"/>
    <bpmn:boundaryEvent id="Boundary_1" attachedToRef="Task_1" cancelActivity="false">
      <bpmn:timerEventDefinition id="TimerEvtDef_1"/>
    </bpmn:boundaryEvent>
    <bpmn:intermediateCatchEvent id="ICatch_1" name="Wait Signal">
      <bpmn:signalEventDefinition id="SigEvtDef_1" signalRef="Sig_1"/>
    </bpmn:intermediateCatchEvent>
    <bpmn:intermediateThrowEvent id="IThrow_1" name="Escalate">
      <bpmn:escalationEventDefinition id="EscEvtDef_1" escalationRef="Esc_1"/>
    </bpmn:intermediateThrowEvent>
    <bpmn:endEvent id="End_1">
      <bpmn:errorEventDefinition id="ErrEvtDef_1" errorRef="Err_1"/>
    </bpmn:endEvent>
    <bpmn:endEvent id="End_2">
      <bpmn:terminateEventDefinition id="TermEvtDef_1"/>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="Task_1"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseIntermediateCatchEvent(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	events := GetTypedElements[IntermediateCatchEvent](bmi.ModelInstance)
	require.Len(t, events, 1)
	assert.Equal(t, "ICatch_1", events[0].GetId())
	assert.Equal(t, "Wait Signal", events[0].GetName())
}

func TestParseIntermediateThrowEvent(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	events := GetTypedElements[IntermediateThrowEvent](bmi.ModelInstance)
	require.Len(t, events, 1)
	assert.Equal(t, "IThrow_1", events[0].GetId())
	assert.Equal(t, "Escalate", events[0].GetName())
}

func TestParseBoundaryEvent(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	events := GetTypedElements[BoundaryEvent](bmi.ModelInstance)
	require.Len(t, events, 1)
	assert.Equal(t, "Boundary_1", events[0].GetId())
	assert.False(t, events[0].IsCancelActivity())
	assert.Equal(t, "Task_1", events[0].GetAttachedToRef())

	attached := events[0].GetAttachedTo()
	require.NotNil(t, attached)
	assert.Equal(t, "Task_1", attached.GetId())
}

func TestParseEventDefinitions(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	msgDefs := GetTypedElements[MessageEventDefinition](bmi.ModelInstance)
	require.Len(t, msgDefs, 1)
	assert.Equal(t, "Msg_1", msgDefs[0].GetMessageRef())

	timerDefs := GetTypedElements[TimerEventDefinition](bmi.ModelInstance)
	require.Len(t, timerDefs, 1)

	sigDefs := GetTypedElements[SignalEventDefinition](bmi.ModelInstance)
	require.Len(t, sigDefs, 1)
	assert.Equal(t, "Sig_1", sigDefs[0].GetSignalRef())

	errDefs := GetTypedElements[ErrorEventDefinition](bmi.ModelInstance)
	require.Len(t, errDefs, 1)
	assert.Equal(t, "Err_1", errDefs[0].GetErrorRef())

	escDefs := GetTypedElements[EscalationEventDefinition](bmi.ModelInstance)
	require.Len(t, escDefs, 1)
	assert.Equal(t, "Esc_1", escDefs[0].GetEscalationRef())

	termDefs := GetTypedElements[TerminateEventDefinition](bmi.ModelInstance)
	require.Len(t, termDefs, 1)
}

func TestParseGlobalElements(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	msgs := GetTypedElements[Message](bmi.ModelInstance)
	require.Len(t, msgs, 1)
	assert.Equal(t, "Msg_1", msgs[0].GetId())
	assert.Equal(t, "OrderReceived", msgs[0].GetName())

	sigs := GetTypedElements[Signal](bmi.ModelInstance)
	require.Len(t, sigs, 1)
	assert.Equal(t, "Sig_1", sigs[0].GetId())
	assert.Equal(t, "AllDone", sigs[0].GetName())

	errs := GetTypedElements[Error](bmi.ModelInstance)
	require.Len(t, errs, 1)
	assert.Equal(t, "Err_1", errs[0].GetId())
	assert.Equal(t, "ValidationError", errs[0].GetName())
	assert.Equal(t, "ERR_001", errs[0].GetErrorCode())

	escs := GetTypedElements[Escalation](bmi.ModelInstance)
	require.Len(t, escs, 1)
	assert.Equal(t, "Esc_1", escs[0].GetId())
	assert.Equal(t, "ESC_HIGH", escs[0].GetEscalationCode())
}

const testBpmnWithCollaboration = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:collaboration id="Collab_1" name="Order Process">
    <bpmn:participant id="Part_1" name="Customer" processRef="Process_1"/>
    <bpmn:participant id="Part_2" name="Supplier" processRef="Process_2"/>
    <bpmn:messageFlow id="MsgFlow_1" name="Order" sourceRef="Part_1" targetRef="Part_2"/>
  </bpmn:collaboration>
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1"/>
    <bpmn:endEvent id="End_1"/>
  </bpmn:process>
  <bpmn:process id="Process_2" isExecutable="false">
    <bpmn:startEvent id="Start_2"/>
    <bpmn:endEvent id="End_2"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseCollaboration(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithCollaboration)
	require.NoError(t, err)

	collabs := GetTypedElements[Collaboration](bmi.ModelInstance)
	require.Len(t, collabs, 1)
	assert.Equal(t, "Collab_1", collabs[0].GetId())
	assert.Equal(t, "Order Process", collabs[0].GetName())
}

func TestParseParticipants(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithCollaboration)
	require.NoError(t, err)

	parts := GetTypedElements[Participant](bmi.ModelInstance)
	require.Len(t, parts, 2)

	// Find by ID
	p1, ok := GetTypedElementById[Participant](bmi.ModelInstance, "Part_1")
	require.True(t, ok)
	assert.Equal(t, "Customer", p1.GetName())
	assert.Equal(t, "Process_1", p1.GetProcessRef())
}

func TestParseMessageFlow(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithCollaboration)
	require.NoError(t, err)

	flows := GetTypedElements[MessageFlow](bmi.ModelInstance)
	require.Len(t, flows, 1)
	assert.Equal(t, "MsgFlow_1", flows[0].GetId())
	assert.Equal(t, "Order", flows[0].GetName())
	assert.Equal(t, "Part_1", flows[0].GetSourceRef())
	assert.Equal(t, "Part_2", flows[0].GetTargetRef())
}

const testBpmnWithSubProcess = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:subProcess id="Sub_1" name="SubProcess" triggeredByEvent="true">
      <bpmn:startEvent id="SubStart_1"/>
      <bpmn:endEvent id="SubEnd_1"/>
    </bpmn:subProcess>
    <bpmn:callActivity id="Call_1" name="Call External" calledElement="ExternalProcess"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseSubProcess(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithSubProcess)
	require.NoError(t, err)

	subs := GetTypedElements[SubProcess](bmi.ModelInstance)
	require.Len(t, subs, 1)
	assert.Equal(t, "Sub_1", subs[0].GetId())
	assert.Equal(t, "SubProcess", subs[0].GetName())
	assert.True(t, subs[0].IsTriggeredByEvent())

	flowElems := subs[0].GetFlowElements()
	assert.Len(t, flowElems, 2)
}

func TestParseCallActivity(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithSubProcess)
	require.NoError(t, err)

	calls := GetTypedElements[CallActivity](bmi.ModelInstance)
	require.Len(t, calls, 1)
	assert.Equal(t, "Call_1", calls[0].GetId())
	assert.Equal(t, "Call External", calls[0].GetName())
	assert.Equal(t, "ExternalProcess", calls[0].GetCalledElement())
}

const testBpmnWithMultiInstance = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:serviceTask id="Task_1" name="Process Items">
      <bpmn:multiInstanceLoopCharacteristics id="MI_1" isSequential="true"/>
    </bpmn:serviceTask>
  </bpmn:process>
</bpmn:definitions>`

func TestParseMultiInstance(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithMultiInstance)
	require.NoError(t, err)

	mis := GetTypedElements[MultiInstanceLoopCharacteristics](bmi.ModelInstance)
	require.Len(t, mis, 1)
	assert.Equal(t, "MI_1", mis[0].GetId())
	assert.True(t, mis[0].IsSequential())
}

const testBpmnWithArtifacts = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1"/>
    <bpmn:textAnnotation id="Ann_1">
      <bpmn:text>Important note</bpmn:text>
    </bpmn:textAnnotation>
    <bpmn:association id="Assoc_1" sourceRef="Start_1" targetRef="Ann_1" associationDirection="One"/>
  </bpmn:process>
</bpmn:definitions>`

func TestParseTextAnnotation(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithArtifacts)
	require.NoError(t, err)

	anns := GetTypedElements[TextAnnotation](bmi.ModelInstance)
	require.Len(t, anns, 1)
	assert.Equal(t, "Ann_1", anns[0].GetId())
	assert.Equal(t, "Important note", anns[0].GetText())
}

func TestParseAssociation(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithArtifacts)
	require.NoError(t, err)

	assocs := GetTypedElements[Association](bmi.ModelInstance)
	require.Len(t, assocs, 1)
	assert.Equal(t, "Assoc_1", assocs[0].GetId())
	assert.Equal(t, "Start_1", assocs[0].GetSourceRef())
	assert.Equal(t, "Ann_1", assocs[0].GetTargetRef())
	assert.Equal(t, "One", assocs[0].GetAssociationDirection())
}

func TestRoundTripWithAllTypes(t *testing.T) {
	bmi, err := ReadFromString(testBpmnWithEvents)
	require.NoError(t, err)

	xml, err := ConvertToString(bmi)
	require.NoError(t, err)

	bmi2, err := ReadFromString(xml)
	require.NoError(t, err)

	// Verify some elements survived the round trip
	assert.Len(t, GetTypedElements[ServiceTask](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[BoundaryEvent](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[IntermediateCatchEvent](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[IntermediateThrowEvent](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[Message](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[Signal](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[Error](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[Escalation](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[MessageEventDefinition](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[TimerEventDefinition](bmi2.ModelInstance), 1)
	assert.Len(t, GetTypedElements[TerminateEventDefinition](bmi2.ModelInstance), 1)
}
