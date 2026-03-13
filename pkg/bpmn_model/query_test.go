package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const graphBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                  id="Definitions_1"
                  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="start">
      <bpmn:outgoing>f1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:exclusiveGateway id="gw">
      <bpmn:incoming>f1</bpmn:incoming>
      <bpmn:outgoing>f2</bpmn:outgoing>
      <bpmn:outgoing>f3</bpmn:outgoing>
    </bpmn:exclusiveGateway>
    <bpmn:serviceTask id="svc1" name="Service A">
      <bpmn:incoming>f2</bpmn:incoming>
      <bpmn:outgoing>f4</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:userTask id="user1" name="Review">
      <bpmn:incoming>f3</bpmn:incoming>
      <bpmn:outgoing>f5</bpmn:outgoing>
    </bpmn:userTask>
    <bpmn:endEvent id="end">
      <bpmn:incoming>f4</bpmn:incoming>
      <bpmn:incoming>f5</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="f1" sourceRef="start" targetRef="gw"/>
    <bpmn:sequenceFlow id="f2" sourceRef="gw" targetRef="svc1">
      <bpmn:conditionExpression xsi:type="bpmn:tFormalExpression">=isValid</bpmn:conditionExpression>
    </bpmn:sequenceFlow>
    <bpmn:sequenceFlow id="f3" sourceRef="gw" targetRef="user1"/>
    <bpmn:sequenceFlow id="f4" sourceRef="svc1" targetRef="end"/>
    <bpmn:sequenceFlow id="f5" sourceRef="user1" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>`

func TestQuery_SucceedingNodes(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	gw, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "gw")
	require.True(t, ok)

	succeeding := SucceedingNodes(gw)
	assert.Equal(t, 2, succeeding.Count())

	nodes := succeeding.List()
	ids := make(map[string]bool)
	for _, n := range nodes {
		ids[n.GetId()] = true
	}
	assert.True(t, ids["svc1"])
	assert.True(t, ids["user1"])
}

func TestQuery_PreviousNodes(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	end, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "end")
	require.True(t, ok)

	previous := PreviousNodes(end)
	assert.Equal(t, 2, previous.Count())

	ids := make(map[string]bool)
	for _, n := range previous.List() {
		ids[n.GetId()] = true
	}
	assert.True(t, ids["svc1"])
	assert.True(t, ids["user1"])
}

func TestQuery_SingleResult(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	start, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "start")
	require.True(t, ok)

	// start has exactly 1 succeeding node
	result, err := SucceedingNodes(start).SingleResult()
	require.NoError(t, err)
	assert.Equal(t, "gw", result.GetId())
}

func TestQuery_SingleResult_Error(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	gw, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "gw")
	require.True(t, ok)

	// gw has 2 succeeding nodes
	_, err = SucceedingNodes(gw).SingleResult()
	assert.Error(t, err)
}

func TestQuery_FilterByType(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	gw, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "gw")
	require.True(t, ok)

	// Filter succeeding nodes to only ServiceTask
	serviceTasks := FilterByType[ServiceTask](SucceedingNodes(gw))
	assert.Equal(t, 1, serviceTasks.Count())
	assert.Equal(t, "svc1", serviceTasks.List()[0].GetId())

	// Filter to UserTask
	userTasks := FilterByType[UserTask](SucceedingNodes(gw))
	assert.Equal(t, 1, userTasks.Count())
	assert.Equal(t, "user1", userTasks.List()[0].GetId())
}

func TestQuery_Filter(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	gw, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "gw")
	require.True(t, ok)

	// Filter succeeding nodes by name
	filtered := SucceedingNodes(gw).Filter(func(n FlowNode) bool {
		return n.GetName() == "Service A"
	})
	assert.Equal(t, 1, filtered.Count())
	assert.Equal(t, "svc1", filtered.List()[0].GetId())
}

func TestQuery_EmptyResult(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	end, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "end")
	require.True(t, ok)

	// end has no succeeding nodes
	succeeding := SucceedingNodes(end)
	assert.Equal(t, 0, succeeding.Count())

	_, err = succeeding.SingleResult()
	assert.Error(t, err)
}

func TestQuery_Processes(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	procs := QueryProcesses(bmi)
	assert.Equal(t, 1, procs.Count())
	assert.Equal(t, "Process_1", procs.List()[0].GetId())
}

func TestQuery_ChainNavigation(t *testing.T) {
	bmi, err := ReadFromString(graphBPMN)
	require.NoError(t, err)

	// Navigate: start -> gw -> svc1 -> end
	start, ok := GetTypedElementById[FlowNode](bmi.ModelInstance, "start")
	require.True(t, ok)

	gw, err := SucceedingNodes(start).SingleResult()
	require.NoError(t, err)
	assert.Equal(t, "gw", gw.GetId())

	svcTasks := FilterByType[ServiceTask](SucceedingNodes(gw))
	svc, err := svcTasks.SingleResult()
	require.NoError(t, err)
	assert.Equal(t, "svc1", svc.GetId())

	endNode, err := SucceedingNodes(svc).SingleResult()
	require.NoError(t, err)
	assert.Equal(t, "end", endNode.GetId())
}
