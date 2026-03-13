package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1">
      <bpmn:outgoing>Flow_1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:endEvent id="EndEvent_1">
      <bpmn:incoming>Flow_1</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="StartEvent_1" targetRef="EndEvent_1"/>
  </bpmn:process>
</bpmn:definitions>`

const bpmnWithCondition = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1">
      <bpmn:outgoing>Flow_1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:endEvent id="End_1">
      <bpmn:incoming>Flow_2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:endEvent id="End_2">
      <bpmn:incoming>Flow_3</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="End_1"/>
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Start_1" targetRef="End_1">
      <bpmn:conditionExpression xsi:type="bpmn:tFormalExpression">=isValid</bpmn:conditionExpression>
    </bpmn:sequenceFlow>
    <bpmn:sequenceFlow id="Flow_3" sourceRef="Start_1" targetRef="End_2"/>
  </bpmn:process>
</bpmn:definitions>`

func TestReadFromString_MinimalBPMN(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)
	require.NotNil(t, bmi)

	// Definitions
	def := bmi.GetDefinitions()
	require.NotNil(t, def)
	assert.Equal(t, "Definitions_1", def.GetId())
	assert.Equal(t, "http://bpmn.io/schema/bpmn", def.GetTargetNamespace())
}

func TestReadFromString_Process(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	// Find process
	processes := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, processes, 1)

	proc := processes[0]
	assert.Equal(t, "Process_1", proc.GetId())
	assert.True(t, proc.IsExecutable())
}

func TestReadFromString_FlowElements(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	proc := GetTypedElements[Process](bmi.ModelInstance)[0]
	flowElements := proc.GetFlowElements()
	assert.Len(t, flowElements, 3) // startEvent, endEvent, sequenceFlow
}

func TestReadFromString_StartEvent(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	startEvents := GetTypedElements[StartEvent](bmi.ModelInstance)
	require.Len(t, startEvents, 1)

	start := startEvents[0]
	assert.Equal(t, "StartEvent_1", start.GetId())
	assert.True(t, start.IsInterrupting()) // default true
}

func TestReadFromString_EndEvent(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	endEvents := GetTypedElements[EndEvent](bmi.ModelInstance)
	require.Len(t, endEvents, 1)
	assert.Equal(t, "EndEvent_1", endEvents[0].GetId())
}

func TestReadFromString_SequenceFlow(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	flows := GetTypedElements[SequenceFlow](bmi.ModelInstance)
	require.Len(t, flows, 1)

	flow := flows[0]
	assert.Equal(t, "Flow_1", flow.GetId())
	assert.Equal(t, "StartEvent_1", flow.GetSourceRef())
	assert.Equal(t, "EndEvent_1", flow.GetTargetRef())
}

func TestSequenceFlow_ResolveReferences(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	flow := GetTypedElements[SequenceFlow](bmi.ModelInstance)[0]

	source := flow.GetSource()
	require.NotNil(t, source)
	assert.Equal(t, "StartEvent_1", source.GetId())

	target := flow.GetTarget()
	require.NotNil(t, target)
	assert.Equal(t, "EndEvent_1", target.GetId())
}

func TestFlowNode_IncomingOutgoing(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	start := GetTypedElements[StartEvent](bmi.ModelInstance)[0]
	outgoing := start.GetOutgoingSequenceFlows()
	require.Len(t, outgoing, 1)
	assert.Equal(t, "Flow_1", outgoing[0].GetId())
	assert.Empty(t, start.GetIncomingSequenceFlows())

	end := GetTypedElements[EndEvent](bmi.ModelInstance)[0]
	incoming := end.GetIncomingSequenceFlows()
	require.Len(t, incoming, 1)
	assert.Equal(t, "Flow_1", incoming[0].GetId())
	assert.Empty(t, end.GetOutgoingSequenceFlows())
}

func TestGetElementById(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	start, ok := GetTypedElementById[StartEvent](bmi.ModelInstance, "StartEvent_1")
	require.True(t, ok)
	assert.Equal(t, "StartEvent_1", start.GetId())

	proc, ok := GetTypedElementById[Process](bmi.ModelInstance, "Process_1")
	require.True(t, ok)
	assert.Equal(t, "Process_1", proc.GetId())

	_, ok = GetTypedElementById[StartEvent](bmi.ModelInstance, "nonexistent")
	assert.False(t, ok)
}

func TestConditionExpression(t *testing.T) {
	bmi, err := ReadFromString(bpmnWithCondition)
	require.NoError(t, err)

	flow2, ok := GetTypedElementById[SequenceFlow](bmi.ModelInstance, "Flow_2")
	require.True(t, ok)

	cond := flow2.GetConditionExpression()
	require.NotNil(t, cond)
	assert.Equal(t, "=isValid", cond.GetTextContent())

	// Flow without condition
	flow1, ok := GetTypedElementById[SequenceFlow](bmi.ModelInstance, "Flow_1")
	require.True(t, ok)
	assert.Nil(t, flow1.GetConditionExpression())
}

func TestCreateProcess(t *testing.T) {
	bmi := CreateProcess().Done()
	require.NotNil(t, bmi)

	def := bmi.GetDefinitions()
	require.NotNil(t, def)
	assert.Equal(t, "http://bpmn.io/schema/bpmn", def.GetTargetNamespace())

	processes := GetTypedElements[Process](bmi.ModelInstance)
	require.Len(t, processes, 1)
	assert.False(t, processes[0].IsExecutable())
}

func TestCreateExecutableProcess(t *testing.T) {
	bmi := CreateExecutableProcess("myProcess").Done()
	require.NotNil(t, bmi)

	proc, ok := GetTypedElementById[Process](bmi.ModelInstance, "myProcess")
	require.True(t, ok)
	assert.True(t, proc.IsExecutable())
	assert.Equal(t, "myProcess", proc.GetId())
}

func TestConvertToString(t *testing.T) {
	bmi := CreateExecutableProcess("test-proc").Done()

	xml, err := ConvertToString(bmi)
	require.NoError(t, err)

	assert.Contains(t, xml, "definitions")
	assert.Contains(t, xml, "test-proc")
	assert.Contains(t, xml, "isExecutable")
	assert.Contains(t, xml, `xmlns:bpmn=`)
}

func TestRoundTrip(t *testing.T) {
	// Parse
	bmi1, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	// Serialize
	xml1, err := ConvertToString(bmi1)
	require.NoError(t, err)

	// Parse again
	bmi2, err := ReadFromString(xml1)
	require.NoError(t, err)

	// Verify structure preserved
	def := bmi2.GetDefinitions()
	require.NotNil(t, def)
	assert.Equal(t, "Definitions_1", def.GetId())

	proc, ok := GetTypedElementById[Process](bmi2.ModelInstance, "Process_1")
	require.True(t, ok)
	assert.True(t, proc.IsExecutable())

	start, ok := GetTypedElementById[StartEvent](bmi2.ModelInstance, "StartEvent_1")
	require.True(t, ok)
	assert.NotNil(t, start)

	flow, ok := GetTypedElementById[SequenceFlow](bmi2.ModelInstance, "Flow_1")
	require.True(t, ok)
	assert.Equal(t, "StartEvent_1", flow.GetSourceRef())
	assert.Equal(t, "EndEvent_1", flow.GetTargetRef())
}

func TestCreateAndBuildProcess(t *testing.T) {
	bmi := CreateExecutableProcess("order-process").Done()
	proc, _ := GetTypedElementById[Process](bmi.ModelInstance, "order-process")

	// Create start event
	startType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_START_EVENT)
	startInst, err := bmi.NewInstance(startType)
	require.NoError(t, err)
	start := startInst.(StartEvent)
	start.SetId("start")

	// Create end event
	endType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_END_EVENT)
	endInst, err := bmi.NewInstance(endType)
	require.NoError(t, err)
	end := endInst.(EndEvent)
	end.SetId("end")

	// Create sequence flow
	flowType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_SEQUENCE_FLOW)
	flowInst, err := bmi.NewInstance(flowType)
	require.NoError(t, err)
	flow := flowInst.(SequenceFlow)
	flow.SetId("flow1")
	flow.SetSourceRef("start")
	flow.SetTargetRef("end")

	// Add to process
	proc.AddFlowElement(start)
	proc.AddFlowElement(end)
	proc.AddFlowElement(flow)

	// Verify
	assert.Len(t, proc.GetFlowElements(), 3)
	assert.Equal(t, "start", flow.GetSource().GetId())
	assert.Equal(t, "end", flow.GetTarget().GetId())

	// Serialize
	xml, err := ConvertToString(bmi)
	require.NoError(t, err)
	assert.Contains(t, xml, "startEvent")
	assert.Contains(t, xml, "endEvent")
	assert.Contains(t, xml, "sequenceFlow")
	assert.Contains(t, xml, `sourceRef="start"`)
	assert.Contains(t, xml, `targetRef="end"`)
}

func TestMultipleProcesses(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true" name="Main"/>
  <bpmn:process id="Process_2" isExecutable="false" name="Secondary"/>
</bpmn:definitions>`

	bmi, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	processes := GetTypedElements[Process](bmi.ModelInstance)
	assert.Len(t, processes, 2)

	proc1, ok := GetTypedElementById[Process](bmi.ModelInstance, "Process_1")
	require.True(t, ok)
	assert.Equal(t, "Main", proc1.GetName())
	assert.True(t, proc1.IsExecutable())

	proc2, ok := GetTypedElementById[Process](bmi.ModelInstance, "Process_2")
	require.True(t, ok)
	assert.Equal(t, "Secondary", proc2.GetName())
	assert.False(t, proc2.IsExecutable())
}

func TestDefinitions_Attributes(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="Defs_1"
                  name="My Model"
                  targetNamespace="http://example.com"
                  exporter="Camunda Modeler"
                  exporterVersion="5.0.0">
</bpmn:definitions>`

	bmi, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	def := bmi.GetDefinitions()
	require.NotNil(t, def)
	assert.Equal(t, "Defs_1", def.GetId())
	assert.Equal(t, "My Model", def.GetName())
	assert.Equal(t, "http://example.com", def.GetTargetNamespace())
	assert.Equal(t, "Camunda Modeler", def.GetExporter())
	assert.Equal(t, "5.0.0", def.GetExporterVersion())
}

func TestFlowElement_Name(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" id="D1">
  <bpmn:process id="P1" isExecutable="true">
    <bpmn:startEvent id="S1" name="Order Received"/>
    <bpmn:endEvent id="E1" name="Order Completed"/>
  </bpmn:process>
</bpmn:definitions>`

	bmi, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	start, ok := GetTypedElementById[StartEvent](bmi.ModelInstance, "S1")
	require.True(t, ok)
	assert.Equal(t, "Order Received", start.GetName())

	end, ok := GetTypedElementById[EndEvent](bmi.ModelInstance, "E1")
	require.True(t, ok)
	assert.Equal(t, "Order Completed", end.GetName())
}

func TestProcessType_Default(t *testing.T) {
	bmi, err := ReadFromString(minimalBPMN)
	require.NoError(t, err)

	proc := GetTypedElements[Process](bmi.ModelInstance)[0]
	assert.Equal(t, "None", proc.GetProcessType())
	assert.False(t, proc.IsClosed())
}

func TestRoundTrip_CreateProcessAndReparse(t *testing.T) {
	// Create programmatically
	bmi := CreateExecutableProcess("test").Done()
	proc, _ := GetTypedElementById[Process](bmi.ModelInstance, "test")
	proc.SetName("Test Process")

	startType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_START_EVENT)
	startInst, _ := bmi.NewInstance(startType)
	start := startInst.(StartEvent)
	start.SetId("s1")
	start.SetName("Begin")
	proc.AddFlowElement(start)

	// Serialize
	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	// Verify XML contains expected content
	assert.Contains(t, xmlStr, "Test Process")
	assert.Contains(t, xmlStr, "Begin")

	// Re-parse
	bmi2, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	proc2, ok := GetTypedElementById[Process](bmi2.ModelInstance, "test")
	require.True(t, ok)
	assert.Equal(t, "Test Process", proc2.GetName())
	assert.True(t, proc2.IsExecutable())

	start2, ok := GetTypedElementById[StartEvent](bmi2.ModelInstance, "s1")
	require.True(t, ok)
	assert.Equal(t, "Begin", start2.GetName())
}
