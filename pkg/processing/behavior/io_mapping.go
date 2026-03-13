package behavior

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/feel"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// applyInputMappings evaluates ZeebeInput expressions from the BPMN element
// and writes the results as variables in the element's scope.
// Source expressions are evaluated against the parent scope variables.
func applyInputMappings(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, elementId string, elementScopeKey, parentScopeKey uint64, piKey uint64) error {
	inputs, _ := getIoMappings(bmi, elementId)
	if len(inputs) == 0 {
		return nil
	}

	scope, err := loadScope(ctx, s, parentScopeKey)
	if err != nil {
		return fmt.Errorf("failed to load parent scope for input mapping: %w", err)
	}

	for _, input := range inputs {
		source := input.GetSource()
		target := input.GetTarget()

		result, err := feel.Eval(source, scope)
		if err != nil {
			return fmt.Errorf("input mapping error (source=%q, target=%q): %w", source, target, err)
		}

		valBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("input mapping marshal error: %w", err)
		}

		if err := s.Variables().Create(ctx, &storage.Variable{
			ProcessInstanceKey: piKey,
			ScopeKey:           elementScopeKey,
			Name:               target,
			Value:              valBytes,
		}); err != nil {
			return fmt.Errorf("input mapping create variable error: %w", err)
		}
	}

	return nil
}

// applyOutputMappings evaluates ZeebeOutput expressions from the BPMN element
// and writes the results as variables in the parent scope.
// Source expressions are evaluated against the element's scope variables.
func applyOutputMappings(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, elementId string, elementScopeKey, parentScopeKey uint64, piKey uint64) error {
	_, outputs := getIoMappings(bmi, elementId)
	if len(outputs) == 0 {
		return nil
	}

	scope, err := loadScope(ctx, s, elementScopeKey)
	if err != nil {
		return fmt.Errorf("failed to load element scope for output mapping: %w", err)
	}

	for _, output := range outputs {
		source := output.GetSource()
		target := output.GetTarget()

		result, err := feel.Eval(source, scope)
		if err != nil {
			return fmt.Errorf("output mapping error (source=%q, target=%q): %w", source, target, err)
		}

		valBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("output mapping marshal error: %w", err)
		}

		// Try to update existing variable, or create new one
		existing, err := s.Variables().FindByName(ctx, parentScopeKey, target)
		if err != nil {
			return err
		}
		if existing != nil {
			if err := s.Variables().Update(ctx, existing.Key, valBytes); err != nil {
				return fmt.Errorf("output mapping update variable error: %w", err)
			}
		} else {
			if err := s.Variables().Create(ctx, &storage.Variable{
				ProcessInstanceKey: piKey,
				ScopeKey:           parentScopeKey,
				Name:               target,
				Value:              valBytes,
			}); err != nil {
				return fmt.Errorf("output mapping create variable error: %w", err)
			}
		}
	}

	return nil
}

// getIoMappings extracts ZeebeInput and ZeebeOutput from a BPMN element.
func getIoMappings(bmi *bpmn_model.BpmnModelInstance, elementId string) ([]bpmn_model.ZeebeInput, []bpmn_model.ZeebeOutput) {
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() != elementId {
			continue
		}
		be, ok := fn.(bpmn_model.BaseElement)
		if !ok {
			return nil, nil
		}
		ioMapping, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeIoMapping](be)
		if !found {
			return nil, nil
		}
		return ioMapping.GetInputs(), ioMapping.GetOutputs()
	}
	return nil, nil
}
