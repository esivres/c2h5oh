package bpmn_model

import (
	"fmt"
	"strings"
)

// ValidationSeverity indicates the severity of a validation issue.
type ValidationSeverity int

const (
	SeverityError ValidationSeverity = iota
	SeverityWarning
)

func (s ValidationSeverity) String() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

// ValidationResult represents a single validation issue.
type ValidationResult struct {
	ElementId string
	Message   string
	Severity  ValidationSeverity
}

func (r ValidationResult) String() string {
	if r.ElementId != "" {
		return fmt.Sprintf("[%s] %s (element: %s)", r.Severity, r.Message, r.ElementId)
	}
	return fmt.Sprintf("[%s] %s", r.Severity, r.Message)
}

// ValidationResults is a collection of validation results.
type ValidationResults struct {
	Results []ValidationResult
}

// HasErrors returns true if there are any error-level results.
func (vr ValidationResults) HasErrors() bool {
	for _, r := range vr.Results {
		if r.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warning-level results.
func (vr ValidationResults) HasWarnings() bool {
	for _, r := range vr.Results {
		if r.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// GetErrors returns only error-level results.
func (vr ValidationResults) GetErrors() []ValidationResult {
	var result []ValidationResult
	for _, r := range vr.Results {
		if r.Severity == SeverityError {
			result = append(result, r)
		}
	}
	return result
}

// GetWarnings returns only warning-level results.
func (vr ValidationResults) GetWarnings() []ValidationResult {
	var result []ValidationResult
	for _, r := range vr.Results {
		if r.Severity == SeverityWarning {
			result = append(result, r)
		}
	}
	return result
}

// String returns a formatted summary of all results.
func (vr ValidationResults) String() string {
	if len(vr.Results) == 0 {
		return "Validation passed: no issues found"
	}
	var sb strings.Builder
	errors := vr.GetErrors()
	warnings := vr.GetWarnings()
	fmt.Fprintf(&sb, "Validation: %d error(s), %d warning(s)\n", len(errors), len(warnings))
	for _, r := range vr.Results {
		fmt.Fprintf(&sb, "  %s\n", r)
	}
	return sb.String()
}

// Validator validates a BPMN model instance.
type Validator interface {
	Validate(bmi *BpmnModelInstance) []ValidationResult
}

// Validate runs all default validators against the model.
func Validate(bmi *BpmnModelInstance) ValidationResults {
	return ValidateWithValidators(bmi, DefaultValidators()...)
}

// ValidateWithValidators runs specified validators against the model.
func ValidateWithValidators(bmi *BpmnModelInstance, validators ...Validator) ValidationResults {
	var results []ValidationResult
	for _, v := range validators {
		results = append(results, v.Validate(bmi)...)
	}
	return ValidationResults{Results: results}
}

// DefaultValidators returns the standard set of BPMN validators.
func DefaultValidators() []Validator {
	return []Validator{
		&StartEventValidator{},
		&EndEventValidator{},
		&SequenceFlowValidator{},
		&ExclusiveGatewayValidator{},
	}
}

// ZeebeValidators returns validators specific to Zeebe runtime.
func ZeebeValidators() []Validator {
	return append(DefaultValidators(),
		&ServiceTaskZeebeValidator{},
		&CallActivityValidator{},
	)
}

// --- Validators ---

// StartEventValidator checks that each process has at least one start event.
type StartEventValidator struct{}

func (v *StartEventValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, proc := range GetTypedElements[Process](bmi.ModelInstance) {
		hasStart := false
		for _, fe := range proc.GetFlowElements() {
			if _, ok := fe.(StartEvent); ok {
				hasStart = true
				break
			}
		}
		if !hasStart {
			results = append(results, ValidationResult{
				ElementId: proc.GetId(),
				Message:   "Process has no start event",
				Severity:  SeverityError,
			})
		}
	}
	return results
}

// EndEventValidator checks that each process has at least one end event.
type EndEventValidator struct{}

func (v *EndEventValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, proc := range GetTypedElements[Process](bmi.ModelInstance) {
		hasEnd := false
		for _, fe := range proc.GetFlowElements() {
			if _, ok := fe.(EndEvent); ok {
				hasEnd = true
				break
			}
		}
		if !hasEnd {
			results = append(results, ValidationResult{
				ElementId: proc.GetId(),
				Message:   "Process has no end event",
				Severity:  SeverityError,
			})
		}
	}
	return results
}

// SequenceFlowValidator checks that sequence flow source and target refs exist.
type SequenceFlowValidator struct{}

func (v *SequenceFlowValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, sf := range GetTypedElements[SequenceFlow](bmi.ModelInstance) {
		if sf.GetSourceRef() == "" {
			results = append(results, ValidationResult{
				ElementId: sf.GetId(),
				Message:   "Sequence flow has no sourceRef",
				Severity:  SeverityError,
			})
		} else if sf.GetSource() == nil {
			results = append(results, ValidationResult{
				ElementId: sf.GetId(),
				Message:   fmt.Sprintf("Sequence flow sourceRef '%s' does not exist", sf.GetSourceRef()),
				Severity:  SeverityError,
			})
		}
		if sf.GetTargetRef() == "" {
			results = append(results, ValidationResult{
				ElementId: sf.GetId(),
				Message:   "Sequence flow has no targetRef",
				Severity:  SeverityError,
			})
		} else if sf.GetTarget() == nil {
			results = append(results, ValidationResult{
				ElementId: sf.GetId(),
				Message:   fmt.Sprintf("Sequence flow targetRef '%s' does not exist", sf.GetTargetRef()),
				Severity:  SeverityError,
			})
		}
	}
	return results
}

// ExclusiveGatewayValidator checks that outgoing flows (except default) have conditions.
type ExclusiveGatewayValidator struct{}

func (v *ExclusiveGatewayValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, gw := range GetTypedElements[ExclusiveGateway](bmi.ModelInstance) {
		outgoing := gw.GetOutgoingSequenceFlows()
		if len(outgoing) <= 1 {
			continue
		}
		defaultFlow := gw.GetDefaultFlow()
		for _, sf := range outgoing {
			if defaultFlow != nil && sf.GetId() == defaultFlow.GetId() {
				continue
			}
			if sf.GetConditionExpression() == nil {
				results = append(results, ValidationResult{
					ElementId: sf.GetId(),
					Message:   fmt.Sprintf("Outgoing flow from exclusive gateway '%s' has no condition expression", gw.GetId()),
					Severity:  SeverityError,
				})
			}
		}
	}
	return results
}

// ServiceTaskZeebeValidator checks that Zeebe service tasks have a job type.
type ServiceTaskZeebeValidator struct{}

func (v *ServiceTaskZeebeValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, task := range GetTypedElements[ServiceTask](bmi.ModelInstance) {
		td, ok := GetSingleExtensionElement[ZeebeTaskDefinition](task)
		if !ok || td.GetType() == "" {
			results = append(results, ValidationResult{
				ElementId: task.GetId(),
				Message:   "Zeebe service task has no job type (zeebe:taskDefinition type)",
				Severity:  SeverityError,
			})
		}
	}
	return results
}

// CallActivityValidator checks that call activities reference a called element.
type CallActivityValidator struct{}

func (v *CallActivityValidator) Validate(bmi *BpmnModelInstance) []ValidationResult {
	var results []ValidationResult
	for _, ca := range GetTypedElements[CallActivity](bmi.ModelInstance) {
		hasRef := ca.GetCalledElement() != ""
		if !hasRef {
			if ce, ok := GetSingleExtensionElement[ZeebeCalledElement](ca); ok {
				hasRef = ce.GetProcessId() != ""
			}
		}
		if !hasRef {
			results = append(results, ValidationResult{
				ElementId: ca.GetId(),
				Message:   "Call activity has no called element reference",
				Severity:  SeverityError,
			})
		}
	}
	return results
}
