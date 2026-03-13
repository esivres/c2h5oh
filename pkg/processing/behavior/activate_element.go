package behavior

import (
	"context"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// activateElement handles ActivateElementIntent (phase 1).
//
// 1. Create element instance in storage (state=Activated).
// 2. Apply input mappings (ZeebeInput) if present.
// 3. Check for multi-instance.
// 4. Emit ElementActivatedIntent for phase 2 dispatch.
func activateElement(ctx context.Context, s storage.Store, i *intent.ActivateElementIntent) ([]intent.Intent, error) {
	piRepo := s.ProcessInstances()

	ei := &storage.ElementInstance{
		Key:                  i.Key,
		ProcessInstanceKey:   i.ProcessInstanceKey,
		ProcessDefinitionKey: i.ProcessDefinitionKey,
		ElementId:            i.ElementId,
		ElementType:          i.ElementType,
		FlowScopeKey:         i.FlowScopeKey,
		State:                storage.ElementInstanceActivated,
		CreatedAt:            time.Now(),
	}
	if err := piRepo.CreateElementInstance(ctx, ei); err != nil {
		return nil, err
	}

	// Apply input mappings if BPMN element has ZeebeIoMapping
	if i.ProcessDefinitionKey != 0 {
		def, err := s.ProcessDefinitions().FindByKey(ctx, i.ProcessDefinitionKey)
		if err != nil {
			return nil, err
		}
		if def != nil {
			bmi, err := bpmn_model.ReadFromBytes(def.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to parse BPMN for input mappings: %w", err)
			}

			// Check for multi-instance: if this element is an MI body, spawn children
			if i.MIInputVariable == "" { // not already a child
				mi := getMultiInstanceInfo(bmi, i.ElementId)
				if mi != nil {
					return handleMultiInstanceActivation(ctx, s, i, mi)
				}
			}

			if err := applyInputMappings(ctx, s, bmi, i.ElementId, i.Key, i.FlowScopeKey, i.ProcessInstanceKey); err != nil {
				return nil, err
			}
		}
	}

	// Set MI input variable on element scope (when this is an MI child)
	if i.MIInputVariable != "" && len(i.MIInputValue) > 0 {
		if err := s.Variables().Create(ctx, &storage.Variable{
			Key:                i.Key,
			ProcessInstanceKey: i.ProcessInstanceKey,
			ScopeKey:           i.Key,
			Name:               i.MIInputVariable,
			Value:              i.MIInputValue,
		}); err != nil {
			return nil, fmt.Errorf("failed to set MI input variable: %w", err)
		}
	}

	// Phase 2: emit ElementActivatedIntent for type-specific dispatch
	return []intent.Intent{
		&intent.ElementActivatedIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey:   i.Key,
			ProcessDefinitionKey: i.ProcessDefinitionKey,
			ElementId:            i.ElementId,
			ElementType:          i.ElementType,
			FlowScopeKey:         i.FlowScopeKey,
			JobType:              i.JobType,
		},
	}, nil
}
