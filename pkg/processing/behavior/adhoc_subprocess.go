package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/feel"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// checkAdHocSubProcessCompletion checks if an ad-hoc subprocess should complete
// after a child element finishes. Evaluates completionCondition if present.
// If condition is true → complete (optionally terminate remaining children).
// If no condition → complete when all children are done.
func checkAdHocSubProcessCompletion(
	ctx context.Context, s storage.Store, ei *storage.ElementInstance, piKey uint64,
) ([]intent.Intent, error) {
	// Check if the scope (FlowScopeKey) is an ad-hoc subprocess
	scopeEI, err := s.ProcessInstances().GetElementInstance(ctx, ei.FlowScopeKey)
	if err != nil {
		return nil, err
	}
	if scopeEI == nil || scopeEI.ElementType != "adHocSubProcess" {
		return nil, nil
	}
	if scopeEI.State != storage.ElementInstanceActivated {
		return nil, nil
	}

	// Load BPMN model to read completionCondition and cancelRemainingInstances
	def, err := s.ProcessDefinitions().FindByKey(ctx, scopeEI.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, nil
	}

	bmi, err := bpmn_model.ReadFromBytes(def.Content)
	if err != nil {
		return nil, err
	}

	ah := findAdHocSubProcess(bmi, scopeEI.ElementId)

	// Check active children
	allEIs, err := s.ProcessInstances().FindElementInstancesByProcessInstance(ctx, piKey)
	if err != nil {
		return nil, err
	}

	var activeChildren []uint64
	for _, child := range allEIs {
		if child.FlowScopeKey == ei.FlowScopeKey && child.State == storage.ElementInstanceActivated {
			activeChildren = append(activeChildren, child.Key)
		}
	}

	// Evaluate completionCondition if present
	completionCondition := ""
	cancelRemaining := true
	if ah != nil {
		completionCondition = ah.GetCompletionCondition()
		cancelRemaining = ah.GetCancelRemainingInstances()
	}

	if completionCondition != "" {
		scope, err := loadScope(ctx, s, ei.FlowScopeKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load scope for ad-hoc completion condition: %w", err)
		}

		condResult, err := feel.EvalBool(completionCondition, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate completionCondition %q: %w", completionCondition, err)
		}

		if condResult {
			return completeAdHocSubProcess(ei.FlowScopeKey, piKey, activeChildren, cancelRemaining)
		}

		// Condition is false — don't complete even if all children are done.
		// Instead, re-evaluate activeElementsCollection to potentially activate new elements.
		if len(activeChildren) == 0 {
			return reEvaluateActiveElements(ctx, s, bmi, ah, scopeEI, piKey, allEIs)
		}
		return nil, nil
	}

	// No completionCondition: complete when all children are done
	if len(activeChildren) > 0 {
		return nil, nil
	}

	return []intent.Intent{
		&intent.CompleteElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ElementInstanceKey: ei.FlowScopeKey,
		},
	}, nil
}

// completeAdHocSubProcess generates intents to complete an ad-hoc subprocess,
// optionally terminating remaining active children first.
func completeAdHocSubProcess(scopeKey, piKey uint64, activeChildren []uint64, cancelRemaining bool) ([]intent.Intent, error) {
	var intents []intent.Intent

	if cancelRemaining {
		for _, childKey := range activeChildren {
			intents = append(intents, &intent.TerminateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: piKey,
				},
				ElementInstanceKey: childKey,
			})
		}
	} else if len(activeChildren) > 0 {
		// cancelRemaining=false but there are still active children → wait
		return nil, nil
	}

	intents = append(intents, &intent.CompleteElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: piKey,
		},
		ElementInstanceKey: scopeKey,
	})

	return intents, nil
}

// reEvaluateActiveElements re-evaluates the activeElementsCollection expression
// and activates any new elements that don't already exist in the ad-hoc scope.
func reEvaluateActiveElements(
	ctx context.Context, s storage.Store,
	bmi *bpmn_model.BpmnModelInstance, ah bpmn_model.AdHocSubProcess,
	scopeEI *storage.ElementInstance, piKey uint64,
	allEIs []*storage.ElementInstance,
) ([]intent.Intent, error) {
	if ah == nil {
		return nil, nil
	}

	zAdHoc, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeAdHoc](ah)
	if !found || zAdHoc.GetActiveElementsCollection() == "" {
		return nil, nil
	}

	scope, err := loadScope(ctx, s, scopeEI.Key)
	if err != nil {
		return nil, err
	}

	result, err := feel.Eval(zAdHoc.GetActiveElementsCollection(), scope)
	if err != nil {
		return nil, nil // expression error on re-evaluation → silently skip
	}

	elementIds, ok := toStringSlice(result)
	if !ok || len(elementIds) == 0 {
		return nil, nil
	}

	// Find which elements are already activated or completed in this scope
	existingElements := make(map[string]bool)
	for _, child := range allEIs {
		if child.FlowScopeKey == scopeEI.Key {
			existingElements[child.ElementId] = true
		}
	}

	// Build inner elements map
	innerElements := make(map[string]bpmn_model.FlowNode)
	for _, fe := range ah.GetFlowElements() {
		if fn, ok := fe.(bpmn_model.FlowNode); ok {
			innerElements[fn.GetId()] = fn
		}
	}

	var intents []intent.Intent
	for _, elemId := range elementIds {
		if existingElements[elemId] {
			continue // already exists
		}
		fn, exists := innerElements[elemId]
		if !exists {
			continue
		}
		intents = append(intents, &intent.ActivateElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ProcessDefinitionKey: scopeEI.ProcessDefinitionKey,
			ElementId:            elemId,
			ElementType:          resolveElementType(fn),
			FlowScopeKey:         scopeEI.Key,
			JobType:              resolveJobType(bmi, fn),
		})
	}

	return intents, nil
}

// activateAdHocElements evaluates activeElementsCollection and activates requested elements.
func activateAdHocElements(
	ctx context.Context, s storage.Store,
	bmi *bpmn_model.BpmnModelInstance, ah bpmn_model.AdHocSubProcess,
	flowScopeKey, adHocKey, piKey, pdKey uint64,
) ([]intent.Intent, error) {
	zAdHoc, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeAdHoc](ah)
	if !found || zAdHoc.GetActiveElementsCollection() == "" {
		return nil, nil // no expression → stay activated, wait for external activation
	}

	expr := zAdHoc.GetActiveElementsCollection()

	scope, err := loadScope(ctx, s, flowScopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load scope for ad-hoc subprocess: %w", err)
	}

	result, err := feel.Eval(expr, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate activeElementsCollection %q: %w", expr, err)
	}

	elementIds, ok := toStringSlice(result)
	if !ok {
		return nil, fmt.Errorf("activeElementsCollection %q did not evaluate to a list of strings", expr)
	}

	if len(elementIds) == 0 {
		return nil, nil // empty list → stay activated
	}

	// Build a map of inner flow elements for quick lookup
	innerElements := make(map[string]bpmn_model.FlowNode)
	for _, fe := range ah.GetFlowElements() {
		if fn, ok := fe.(bpmn_model.FlowNode); ok {
			innerElements[fn.GetId()] = fn
		}
	}

	// Activate each requested element
	var intents []intent.Intent
	for _, elemId := range elementIds {
		fn, exists := innerElements[elemId]
		if !exists {
			continue // skip unknown element IDs
		}
		intents = append(intents, &intent.ActivateElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ProcessDefinitionKey: pdKey,
			ElementId:            elemId,
			ElementType:          resolveElementType(fn),
			FlowScopeKey:         adHocKey, // ad-hoc subprocess instance is the scope
			JobType:              resolveJobType(bmi, fn),
		})
	}

	return intents, nil
}

// findAdHocSubProcess finds an AdHocSubProcess by element ID in the BPMN model.
func findAdHocSubProcess(bmi *bpmn_model.BpmnModelInstance, elementId string) bpmn_model.AdHocSubProcess {
	adHocs := bpmn_model.GetTypedElements[bpmn_model.AdHocSubProcess](bmi.ModelInstance)
	for _, ah := range adHocs {
		if ah.GetId() == elementId {
			return ah
		}
	}
	return nil
}

// toStringSlice converts a FEEL result to []string.
func toStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return nil, false
		}
		result = append(result, s)
	}
	return result, true
}
