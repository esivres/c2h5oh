package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Parsing tests ---

const diXML = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
                  xmlns:dc="http://www.omg.org/spec/DD/20100524/DC"
                  xmlns:di="http://www.omg.org/spec/DD/20100524/DI"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="Start_1"/>
    <bpmn:serviceTask id="Task_1" name="Do Work"/>
    <bpmn:endEvent id="End_1"/>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="Task_1"/>
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Task_1" targetRef="End_1"/>
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_1">
    <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="Process_1">
      <bpmndi:BPMNShape id="Start_1_di" bpmnElement="Start_1">
        <dc:Bounds x="179" y="79" width="36" height="36"/>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="Task_1_di" bpmnElement="Task_1">
        <dc:Bounds x="265" y="57" width="100" height="80"/>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="End_1_di" bpmnElement="End_1">
        <dc:Bounds x="415" y="79" width="36" height="36"/>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNEdge id="Flow_1_di" bpmnElement="Flow_1">
        <di:waypoint x="215" y="97"/>
        <di:waypoint x="265" y="97"/>
      </bpmndi:BPMNEdge>
      <bpmndi:BPMNEdge id="Flow_2_di" bpmnElement="Task_1">
        <di:waypoint x="365" y="97"/>
        <di:waypoint x="415" y="97"/>
      </bpmndi:BPMNEdge>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</bpmn:definitions>`

func TestDI_ParseDiagram(t *testing.T) {
	bmi, err := ReadFromString(diXML)
	require.NoError(t, err)

	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)

	diag := diagrams[0]
	assert.Equal(t, "BPMNDiagram_1", diag.GetId())

	plane := diag.GetPlane()
	require.NotNil(t, plane)
	assert.Equal(t, "BPMNPlane_1", plane.GetId())
	assert.Equal(t, "Process_1", plane.GetBpmnElement())
}

func TestDI_ParseShapes(t *testing.T) {
	bmi, err := ReadFromString(diXML)
	require.NoError(t, err)

	shapes := GetTypedElements[BpmnShape](bmi.ModelInstance)
	require.Len(t, shapes, 3)

	// Find start event shape
	var startShape BpmnShape
	for _, s := range shapes {
		if s.GetBpmnElement() == "Start_1" {
			startShape = s
			break
		}
	}
	require.NotNil(t, startShape)
	assert.Equal(t, "Start_1_di", startShape.GetId())

	bounds := startShape.GetBounds()
	require.NotNil(t, bounds)
	assert.InDelta(t, 179.0, bounds.GetX(), 0.01)
	assert.InDelta(t, 79.0, bounds.GetY(), 0.01)
	assert.InDelta(t, 36.0, bounds.GetWidth(), 0.01)
	assert.InDelta(t, 36.0, bounds.GetHeight(), 0.01)
}

func TestDI_ParseEdges(t *testing.T) {
	bmi, err := ReadFromString(diXML)
	require.NoError(t, err)

	edges := GetTypedElements[BpmnEdge](bmi.ModelInstance)
	require.Len(t, edges, 2)

	var flow1Edge BpmnEdge
	for _, e := range edges {
		if e.GetBpmnElement() == "Flow_1" {
			flow1Edge = e
			break
		}
	}
	require.NotNil(t, flow1Edge)

	waypoints := flow1Edge.GetWaypoints()
	require.Len(t, waypoints, 2)
	assert.InDelta(t, 215.0, waypoints[0].GetX(), 0.01)
	assert.InDelta(t, 97.0, waypoints[0].GetY(), 0.01)
	assert.InDelta(t, 265.0, waypoints[1].GetX(), 0.01)
	assert.InDelta(t, 97.0, waypoints[1].GetY(), 0.01)
}

// --- Builder DI generation tests ---

func TestDI_BuilderGeneratesDI(t *testing.T) {
	bmi := CreateExecutableProcess("diProc").
		StartEvent("start").
		ServiceTask("task").Name("Work").
		EndEvent("end").
		Done()

	// Verify DI was generated
	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)

	plane := diagrams[0].GetPlane()
	require.NotNil(t, plane)
	assert.Equal(t, "diProc", plane.GetBpmnElement())

	shapes := plane.GetShapes()
	assert.Len(t, shapes, 3) // start, task, end

	edges := plane.GetEdges()
	assert.Len(t, edges, 2) // start->task, task->end
}

func TestDI_BuilderShapeBounds(t *testing.T) {
	bmi := CreateExecutableProcess("boundsProc").
		StartEvent("start").
		ServiceTask("task").
		EndEvent("end").
		Done()

	shapes := GetTypedElements[BpmnShape](bmi.ModelInstance)

	// Find shapes by bpmnElement
	shapeMap := make(map[string]BpmnShape)
	for _, s := range shapes {
		shapeMap[s.GetBpmnElement()] = s
	}

	// Start event: 36x36
	startBounds := shapeMap["start"].GetBounds()
	require.NotNil(t, startBounds)
	assert.InDelta(t, 36.0, startBounds.GetWidth(), 0.01)
	assert.InDelta(t, 36.0, startBounds.GetHeight(), 0.01)

	// Service task: 100x80
	taskBounds := shapeMap["task"].GetBounds()
	require.NotNil(t, taskBounds)
	assert.InDelta(t, 100.0, taskBounds.GetWidth(), 0.01)
	assert.InDelta(t, 80.0, taskBounds.GetHeight(), 0.01)

	// End event: 36x36
	endBounds := shapeMap["end"].GetBounds()
	require.NotNil(t, endBounds)
	assert.InDelta(t, 36.0, endBounds.GetWidth(), 0.01)
	assert.InDelta(t, 36.0, endBounds.GetHeight(), 0.01)
}

func TestDI_BuilderEdgeWaypoints(t *testing.T) {
	bmi := CreateExecutableProcess("edgeProc").
		StartEvent("start").
		EndEvent("end").
		Done()

	edges := GetTypedElements[BpmnEdge](bmi.ModelInstance)
	require.Len(t, edges, 1)

	waypoints := edges[0].GetWaypoints()
	require.Len(t, waypoints, 2)
	// Start: right center, End: left center
	assert.Less(t, waypoints[0].GetX(), waypoints[1].GetX())
}

func TestDI_XMLOutput(t *testing.T) {
	bmi := CreateExecutableProcess("xmlDiProc").
		StartEvent("start").
		ServiceTask("task").
		EndEvent("end").
		Done()

	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	assert.Contains(t, xmlStr, "BPMNDiagram")
	assert.Contains(t, xmlStr, "BPMNPlane")
	assert.Contains(t, xmlStr, "BPMNShape")
	assert.Contains(t, xmlStr, "BPMNEdge")
	assert.Contains(t, xmlStr, "Bounds")
	assert.Contains(t, xmlStr, "waypoint")
}

func TestDI_RoundTrip(t *testing.T) {
	// Parse existing DI
	bmi, err := ReadFromString(diXML)
	require.NoError(t, err)

	// Serialize
	xmlStr, err := ConvertToString(bmi)
	require.NoError(t, err)

	// Re-parse
	bmi2, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	// Verify DI preserved
	diagrams := GetTypedElements[BpmnDiagram](bmi2.ModelInstance)
	require.Len(t, diagrams, 1)

	plane := diagrams[0].GetPlane()
	require.NotNil(t, plane)

	shapes := plane.GetShapes()
	assert.Len(t, shapes, 3)

	edges := plane.GetEdges()
	assert.Len(t, edges, 2)

	// Verify bounds preserved
	var taskShape BpmnShape
	for _, s := range shapes {
		if s.GetBpmnElement() == "Task_1" {
			taskShape = s
			break
		}
	}
	require.NotNil(t, taskShape)
	bounds := taskShape.GetBounds()
	require.NotNil(t, bounds)
	assert.InDelta(t, 265.0, bounds.GetX(), 0.01)
	assert.InDelta(t, 57.0, bounds.GetY(), 0.01)
	assert.InDelta(t, 100.0, bounds.GetWidth(), 0.01)
	assert.InDelta(t, 80.0, bounds.GetHeight(), 0.01)
}

func TestDI_GatewaySize(t *testing.T) {
	bmi := CreateExecutableProcess("gwDi").
		StartEvent("start").
		ExclusiveGateway("gw").
		EndEvent("end").
		Done()

	shapes := GetTypedElements[BpmnShape](bmi.ModelInstance)
	shapeMap := make(map[string]BpmnShape)
	for _, s := range shapes {
		shapeMap[s.GetBpmnElement()] = s
	}

	gwBounds := shapeMap["gw"].GetBounds()
	require.NotNil(t, gwBounds)
	assert.InDelta(t, 50.0, gwBounds.GetWidth(), 0.01)
	assert.InDelta(t, 50.0, gwBounds.GetHeight(), 0.01)
}

func TestDI_EmptyProcessDone(t *testing.T) {
	bmi := CreateExecutableProcess("empty").Done()

	diagrams := GetTypedElements[BpmnDiagram](bmi.ModelInstance)
	require.Len(t, diagrams, 1)

	plane := diagrams[0].GetPlane()
	require.NotNil(t, plane)
	assert.Equal(t, "empty", plane.GetBpmnElement())
	assert.Empty(t, plane.GetShapes())
	assert.Empty(t, plane.GetEdges())
}
