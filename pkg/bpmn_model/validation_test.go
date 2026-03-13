package bpmn_model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_ValidModel(t *testing.T) {
	bmi := CreateExecutableProcess("valid").
		StartEvent("start").
		ServiceTask("task").
		EndEvent("end").
		Done()

	results := Validate(bmi)
	assert.False(t, results.HasErrors(), results.String())
}

func TestValidation_NoStartEvent(t *testing.T) {
	bmi := CreateExecutableProcess("noStart").Done()
	proc, _ := GetTypedElementById[Process](bmi.ModelInstance, "noStart")

	// Add only end event manually
	endType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_END_EVENT)
	endInst, _ := bmi.NewInstance(endType)
	end := endInst.(EndEvent)
	end.SetId("end")
	proc.AddFlowElement(end)

	results := Validate(bmi)
	assert.True(t, results.HasErrors())
	errors := results.GetErrors()
	found := false
	for _, e := range errors {
		if e.ElementId == "noStart" && e.Message == "Process has no start event" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected 'no start event' error")
}

func TestValidation_NoEndEvent(t *testing.T) {
	bmi := CreateExecutableProcess("noEnd").Done()
	proc, _ := GetTypedElementById[Process](bmi.ModelInstance, "noEnd")

	// Add only start event manually
	startType := bpmnModel.GetTypeByQName(BPMN20_NS, BPMN_ELEMENT_START_EVENT)
	startInst, _ := bmi.NewInstance(startType)
	start := startInst.(StartEvent)
	start.SetId("start")
	proc.AddFlowElement(start)

	results := Validate(bmi)
	assert.True(t, results.HasErrors())
	errors := results.GetErrors()
	found := false
	for _, e := range errors {
		if e.ElementId == "noEnd" && e.Message == "Process has no end event" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected 'no end event' error")
}

func TestValidation_ExclusiveGatewayNoCondition(t *testing.T) {
	// Gateway with 2 outgoing flows but no conditions and no default
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  id="D1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="P1" isExecutable="true">
    <bpmn:startEvent id="start"><bpmn:outgoing>f1</bpmn:outgoing></bpmn:startEvent>
    <bpmn:exclusiveGateway id="gw">
      <bpmn:incoming>f1</bpmn:incoming>
      <bpmn:outgoing>f2</bpmn:outgoing>
      <bpmn:outgoing>f3</bpmn:outgoing>
    </bpmn:exclusiveGateway>
    <bpmn:endEvent id="end1"><bpmn:incoming>f2</bpmn:incoming></bpmn:endEvent>
    <bpmn:endEvent id="end2"><bpmn:incoming>f3</bpmn:incoming></bpmn:endEvent>
    <bpmn:sequenceFlow id="f1" sourceRef="start" targetRef="gw"/>
    <bpmn:sequenceFlow id="f2" sourceRef="gw" targetRef="end1"/>
    <bpmn:sequenceFlow id="f3" sourceRef="gw" targetRef="end2"/>
  </bpmn:process>
</bpmn:definitions>`

	bmi, err := ReadFromString(xmlStr)
	require.NoError(t, err)

	results := Validate(bmi)
	assert.True(t, results.HasErrors())

	// Should have 2 errors (both flows missing conditions)
	gwErrors := 0
	for _, e := range results.GetErrors() {
		if e.Message != "" && (e.ElementId == "f2" || e.ElementId == "f3") {
			gwErrors++
		}
	}
	assert.Equal(t, 2, gwErrors)
}

func TestValidation_ExclusiveGateway_WithDefaultFlow(t *testing.T) {
	bmi := CreateExecutableProcess("gwValid").
		StartEvent("start").
		ExclusiveGateway("gw").
		ConditionExpression("=isValid").
		ServiceTask("task").
		EndEvent("end1").
		MoveToLastGateway().
		DefaultFlow().
		EndEvent("end2").
		Done()

	results := Validate(bmi)
	// Gateway has condition on one flow and default on the other — valid
	gwErrors := 0
	for _, e := range results.GetErrors() {
		if e.Message != "" && (e.ElementId == "gw" || contains(e.Message, "gateway")) {
			gwErrors++
		}
	}
	assert.Equal(t, 0, gwErrors)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestValidation_ZeebeServiceTaskNoJobType(t *testing.T) {
	bmi := CreateExecutableProcess("zeebeVal").
		StartEvent("start").
		ServiceTask("svc"). // No ZeebeJobType!
		EndEvent("end").
		Done()

	results := ValidateWithValidators(bmi, ZeebeValidators()...)
	assert.True(t, results.HasErrors())

	found := false
	for _, e := range results.GetErrors() {
		if e.ElementId == "svc" && containsSubstr(e.Message, "job type") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected zeebe job type error")
}

func TestValidation_ZeebeServiceTaskWithJobType(t *testing.T) {
	bmi := CreateExecutableProcess("zeebeOk").
		StartEvent("start").
		ServiceTask("svc").ZeebeJobType("my-worker").
		EndEvent("end").
		Done()

	results := ValidateWithValidators(bmi, ZeebeValidators()...)
	// Should not have service task zeebe error
	for _, e := range results.GetErrors() {
		assert.NotEqual(t, "svc", e.ElementId, "Service task should not have errors: "+e.Message)
	}
}

func TestValidation_CallActivityNoCalledElement(t *testing.T) {
	bmi := CreateExecutableProcess("callVal").
		StartEvent("start").
		CallActivity("call"). // No called element!
		EndEvent("end").
		Done()

	results := ValidateWithValidators(bmi, ZeebeValidators()...)
	found := false
	for _, e := range results.GetErrors() {
		if e.ElementId == "call" && containsSubstr(e.Message, "called element") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected call activity error")
}

func TestValidation_CallActivityWithZeebeProcessId(t *testing.T) {
	bmi := CreateExecutableProcess("callOk").
		StartEvent("start").
		CallActivity("call").ZeebeProcessId("child").
		EndEvent("end").
		Done()

	results := ValidateWithValidators(bmi, ZeebeValidators()...)
	for _, e := range results.GetErrors() {
		assert.NotEqual(t, "call", e.ElementId, "Call activity should not have errors: "+e.Message)
	}
}

// Custom validator test
type noManualTaskValidator struct{}

func (v *noManualTaskValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, mt := range GetTypedElements[ManualTask](bmi.ModelInstance) {
		results = append(results, ValidationResult{
			ElementId: mt.GetId(),
			Message:   "Manual tasks are not allowed",
			Severity:  SeverityError,
		})
	}
	return results
}

func TestValidation_CustomValidator(t *testing.T) {
	bmi := CreateExecutableProcess("custom").
		StartEvent("start").
		ManualTask("manual").
		EndEvent("end").
		Done()

	results := ValidateWithValidators(bmi, &noManualTaskValidator{})
	assert.True(t, results.HasErrors())
	assert.Equal(t, "manual", results.GetErrors()[0].ElementId)
}

func TestValidation_ResultsFormatting(t *testing.T) {
	results := ValidationResults{
		Results: []ValidationResult{
			{ElementId: "task1", Message: "Missing name", Severity: SeverityWarning},
			{ElementId: "flow1", Message: "No condition", Severity: SeverityError},
		},
	}

	s := results.String()
	assert.Contains(t, s, "1 error(s)")
	assert.Contains(t, s, "1 warning(s)")
	assert.Contains(t, s, "ERROR")
	assert.Contains(t, s, "WARNING")
}
