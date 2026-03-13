package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// createProcessInstance handles CreateProcessInstanceIntent.
//
// 1. Resolve process definition (by key, or by bpmnProcessId + version/latest).
// 2. Create process instance in storage.
// 3. If variables provided — emit SetVariablesIntent.
// 4. Parse BPMN to find start events — emit ActivateElementIntent for each.
func createProcessInstance(ctx context.Context, s storage.Store, i *intent.CreateProcessInstanceIntent) ([]intent.Intent, error) {
	defRepo := s.ProcessDefinitions()
	piRepo := s.ProcessInstances()

	// Resolve process definition
	var def *storage.ProcessDefinition
	var err error

	if i.ProcessDefinitionKey != 0 {
		def, err = defRepo.FindByKey(ctx, i.ProcessDefinitionKey)
	} else if i.Version != 0 {
		def, err = defRepo.FindByProcessIdAndVersion(ctx, i.BpmnProcessId, i.Version)
	} else {
		def, err = defRepo.FindLatestByProcessId(ctx, i.BpmnProcessId)
	}
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, fmt.Errorf("process definition not found: %s", i.BpmnProcessId)
	}

	// Create process instance
	pi := &storage.ProcessInstance{
		Key:                  i.Key,
		ProcessDefinitionKey: def.Key,
		BpmnProcessId:        def.BpmnProcessId,
		ParentKey:            i.ParentKey,
		ParentElementKey:     i.ParentElementKey,
		State:                storage.ProcessInstanceActive,
		CreatedAt:            time.Now(),
	}
	if err := piRepo.CreateInstance(ctx, pi); err != nil {
		return nil, err
	}

	var intents []intent.Intent

	// Set initial variables
	if len(i.Variables) > 0 {
		// Validate JSON
		var check json.RawMessage
		if err := json.Unmarshal(i.Variables, &check); err != nil {
			return nil, fmt.Errorf("invalid variables JSON: %w", err)
		}

		intents = append(intents, &intent.SetVariablesIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.Key,
			},
			ScopeKey:  i.Key, // process instance is the root scope
			Variables: i.Variables,
		})
	}

	// Parse BPMN XML to find start events
	bmi, err := bpmn_model.ReadFromBytes(def.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BPMN: %w", err)
	}

	startEvents := bpmn_model.GetTypedElements[bpmn_model.StartEvent](bmi.ModelInstance)
	for _, se := range startEvents {
		intents = append(intents, &intent.ActivateElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.Key,
			},
			ProcessDefinitionKey: def.Key,
			ElementId:            se.GetId(),
			ElementType:          "startEvent",
			FlowScopeKey:         i.Key,
		})
	}

	return intents, nil
}
