package behavior

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/feel"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// multiInstanceInfo holds the parsed multi-instance configuration for an element.
type multiInstanceInfo struct {
	InputCollection  string
	InputElement     string
	OutputCollection string
	OutputElement    string
	IsSequential     bool
}

// getMultiInstanceInfo checks if a BPMN element has multi-instance loop characteristics.
// Returns nil if not a multi-instance element.
func getMultiInstanceInfo(bmi *bpmn_model.BpmnModelInstance, elementId string) *multiInstanceInfo {
	// Find the element as BaseElement to access extensions
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() != elementId {
			continue
		}

		be, ok := fn.(bpmn_model.BaseElement)
		if !ok {
			return nil
		}

		// Check for ZeebeLoopCharacteristics in extensions
		lc, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeLoopCharacteristics](be)
		if !found {
			return nil
		}

		// Check if there's a MultiInstanceLoopCharacteristics (for isSequential)
		isSequential := false
		miList := bpmn_model.GetTypedElements[bpmn_model.MultiInstanceLoopCharacteristics](bmi.ModelInstance)
		for _, mi := range miList {
			// Check if this MI belongs to our element by checking DOM parent
			if mi.GetDomElement() != nil && mi.GetDomElement().Parent() != nil {
				parentId := mi.GetDomElement().Parent().GetAttribute("id")
				if parentId == elementId {
					isSequential = mi.IsSequential()
					break
				}
			}
		}

		return &multiInstanceInfo{
			IsSequential:     isSequential,
			InputCollection:  lc.GetInputCollection(),
			InputElement:     lc.GetInputElement(),
			OutputCollection: lc.GetOutputCollection(),
			OutputElement:    lc.GetOutputElement(),
		}
	}

	return nil
}

// handleMultiInstanceActivation creates child element instances for a multi-instance element.
// For parallel: creates all children at once. For sequential: creates the first child.
func handleMultiInstanceActivation(
	ctx context.Context,
	s storage.Store,
	i *intent.ActivateElementIntent,
	mi *multiInstanceInfo,
) ([]intent.Intent, error) {
	if mi.InputCollection == "" {
		return nil, fmt.Errorf("multi-instance on %q has no inputCollection", i.ElementId)
	}

	// Evaluate input collection expression
	scope, err := loadScope(ctx, s, i.FlowScopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load scope for multi-instance: %w", err)
	}

	collectionVal, err := feel.Eval(mi.InputCollection, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate inputCollection %q: %w", mi.InputCollection, err)
	}

	items, ok := toSlice(collectionVal)
	if !ok {
		return nil, fmt.Errorf("inputCollection %q did not evaluate to a list", mi.InputCollection)
	}

	if len(items) == 0 {
		// Empty collection → complete immediately
		return []intent.Intent{
			&intent.CompleteElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: i.Key,
			},
		}, nil
	}

	// Store collection metadata as variables on the multi-instance body scope
	collJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MI collection: %w", err)
	}
	countJSON, err := json.Marshal(len(items))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MI count: %w", err)
	}

	var intents []intent.Intent

	// Store the collection and count on the MI body scope
	intents = append(intents, &intent.SetVariablesIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: i.ProcessInstanceKey,
		},
		ScopeKey: i.Key,
		Variables: mustJSON(map[string]any{
			"__mi_collection": json.RawMessage(collJSON),
			"__mi_count":      json.RawMessage(countJSON),
			"__mi_completed":  0,
		}),
	})

	if mi.IsSequential {
		// Sequential: activate only the first child
		intents = append(intents, createMIChildIntent(i, mi, items[0], 0))
	} else {
		// Parallel: activate all children
		for idx, item := range items {
			intents = append(intents, createMIChildIntent(i, mi, item, idx))
		}
	}

	return intents, nil
}

// createMIChildIntent creates an ActivateElementIntent for one multi-instance iteration.
func createMIChildIntent(parent *intent.ActivateElementIntent, mi *multiInstanceInfo, item any, index int) *intent.ActivateElementIntent {
	child := &intent.ActivateElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: parent.ProcessInstanceKey,
		},
		ProcessDefinitionKey: parent.ProcessDefinitionKey,
		ElementId:            parent.ElementId,
		ElementType:          parent.ElementType,
		FlowScopeKey:         parent.Key, // MI body is the scope
		JobType:              parent.JobType,
	}
	// Store the item variable name and value for input mapping at activation
	if mi.InputElement != "" {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			// Fallback to null — should not happen for FEEL-evaluated values
			itemJSON = []byte("null")
		}
		child.MIInputVariable = mi.InputElement
		child.MIInputValue = itemJSON
		child.MIIndex = index
	}
	return child
}

func toSlice(v any) ([]any, bool) {
	if val, ok := v.([]any); ok {
		return val, true
	}
	return nil, false
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
