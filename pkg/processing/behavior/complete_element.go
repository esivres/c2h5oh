package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/feel"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// completeElement handles CompleteElementIntent.
//
// Phase 1: Update element instance state to Completed, apply output mappings.
// Phase 2: Emit ElementCompletedIntent and immediately dispatch to type-specific behavior.
func completeElement(ctx context.Context, s storage.Store, i *intent.CompleteElementIntent) ([]intent.Intent, error) {
	piRepo := s.ProcessInstances()

	// Get element instance to know which element completed
	ei, err := piRepo.GetElementInstance(ctx, i.ElementInstanceKey)
	if err != nil {
		return nil, err
	}
	if ei == nil {
		return nil, fmt.Errorf("element instance not found: %d", i.ElementInstanceKey)
	}

	// Update element instance state
	if err := piRepo.UpdateElementInstanceState(ctx, i.ElementInstanceKey, storage.ElementInstanceCompleted); err != nil {
		return nil, err
	}

	// Load BPMN for output mappings
	bmi, err := loadBPMN(ctx, s, ei.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}

	// Apply output mappings if present
	if err := applyOutputMappings(ctx, s, bmi, ei.ElementId, i.ElementInstanceKey, ei.FlowScopeKey, ei.ProcessInstanceKey); err != nil {
		return nil, err
	}

	// Return ElementCompletedIntent for phase 2 dispatch via the processor/registry
	return []intent.Intent{
		&intent.ElementCompletedIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey:   i.ElementInstanceKey,
			ProcessDefinitionKey: ei.ProcessDefinitionKey,
			ElementId:            ei.ElementId,
			ElementType:          ei.ElementType,
			FlowScopeKey:         ei.FlowScopeKey,
		},
	}, nil
}

// resolveExclusiveGateway evaluates conditions on outgoing flows and returns the target
// of the first flow whose condition is true, or the default flow target.
// Returns nil if no flow can be taken.
func resolveExclusiveGateway(ctx context.Context, s storage.Store, gw bpmn_model.ExclusiveGateway, scopeKey uint64) (bpmn_model.FlowNode, error) {
	outgoing := gw.GetOutgoingSequenceFlows()
	if len(outgoing) == 0 {
		return nil, nil
	}

	// Single outgoing flow — take it unconditionally
	if len(outgoing) == 1 {
		return outgoing[0].GetTarget(), nil
	}

	defaultFlow := gw.GetDefaultFlow()
	scope, err := loadScope(ctx, s, scopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load variables for gateway scope %d: %w", scopeKey, err)
	}

	var defaultTarget bpmn_model.FlowNode

	for _, sf := range outgoing {
		target := sf.GetTarget()
		if target == nil {
			continue
		}

		// Check if this is the default flow
		if defaultFlow != nil && sf.GetId() == defaultFlow.GetId() {
			defaultTarget = target
			continue
		}

		// Evaluate condition
		cond := sf.GetConditionExpression()
		if cond == nil {
			// No condition and not default — treat as unconditionally true
			return target, nil
		}

		expr := cond.GetTextContent()
		if expr == "" {
			return target, nil
		}

		result, err := feel.EvalBool(expr, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition on flow %q: %w", sf.GetId(), err)
		}
		if result {
			return target, nil
		}
	}

	// No condition was true — take default flow
	if defaultTarget != nil {
		return defaultTarget, nil
	}

	return nil, nil
}

// resolveInclusiveGateway evaluates conditions on all outgoing flows and returns targets
// for ALL flows whose conditions are true. If none are true, returns the default flow target.
func resolveInclusiveGateway(ctx context.Context, s storage.Store, gw bpmn_model.InclusiveGateway, scopeKey uint64) ([]bpmn_model.FlowNode, error) {
	outgoing := gw.GetOutgoingSequenceFlows()
	if len(outgoing) == 0 {
		return nil, nil
	}

	if len(outgoing) == 1 {
		if t := outgoing[0].GetTarget(); t != nil {
			return []bpmn_model.FlowNode{t}, nil
		}
		return nil, nil
	}

	defaultFlow := gw.GetDefaultFlow()
	scope, err := loadScope(ctx, s, scopeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load variables for inclusive gateway scope %d: %w", scopeKey, err)
	}

	var targets []bpmn_model.FlowNode
	var defaultTarget bpmn_model.FlowNode

	for _, sf := range outgoing {
		target := sf.GetTarget()
		if target == nil {
			continue
		}

		if defaultFlow != nil && sf.GetId() == defaultFlow.GetId() {
			defaultTarget = target
			continue
		}

		cond := sf.GetConditionExpression()
		if cond == nil {
			targets = append(targets, target)
			continue
		}

		expr := cond.GetTextContent()
		if expr == "" {
			targets = append(targets, target)
			continue
		}

		result, err := feel.EvalBool(expr, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition on flow %q: %w", sf.GetId(), err)
		}
		if result {
			targets = append(targets, target)
		}
	}

	// If no conditional flow was true, take the default
	if len(targets) == 0 && defaultTarget != nil {
		targets = append(targets, defaultTarget)
	}

	return targets, nil
}

// handleErrorEndEvent checks if the end event has an errorEventDefinition.
// If so, propagates the error up to find a matching boundary error event on the parent scope.
// Returns nil intents if this is NOT an error end event (caller should continue normal logic).
func handleErrorEndEvent(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance, piKey uint64) ([]intent.Intent, error) {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"

	// Check if this end event has an errorEventDefinition
	endEvents := bpmn_model.GetTypedElements[bpmn_model.EndEvent](bmi.ModelInstance)
	var errorCode string
	var isErrorEnd bool

	for _, ee := range endEvents {
		if ee.GetId() != ei.ElementId {
			continue
		}
		dom := ee.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			break
		}
		errorDefs := dom.GetChildElementsByNS(bpmnNS, "errorEventDefinition")
		if len(errorDefs) == 0 {
			break
		}
		isErrorEnd = true
		// Get errorRef → resolve error code
		errorRef := errorDefs[0].GetAttribute("errorRef")
		if errorRef != "" {
			errors := bpmn_model.GetTypedElements[bpmn_model.Error](bmi.ModelInstance)
			for _, e := range errors {
				if e.GetId() == errorRef {
					errorCode = e.GetErrorCode()
					break
				}
			}
			if errorCode == "" {
				errorCode = errorRef
			}
		}
		break
	}

	if !isErrorEnd {
		return nil, nil // not an error end event
	}

	// Find the scope element (subprocess or process) that contains this end event.
	// Walk up: flowScopeKey → find its parent scope and look for boundary error event.
	scopeKey := ei.FlowScopeKey

	// If inside a subprocess, find the subprocess element to get its elementId
	if scopeKey != 0 && scopeKey != piKey {
		scopeEI, err := s.ProcessInstances().GetElementInstance(ctx, scopeKey)
		if err != nil {
			return nil, err
		}
		if scopeEI != nil {
			// Look for boundary error event attached to this subprocess
			boundaryId := findBoundaryErrorEvent(bmi, scopeEI.ElementId, errorCode)
			if boundaryId != "" {
				var intents []intent.Intent
				// Terminate all elements inside the subprocess scope
				termIntents, err := terminateScopeElements(ctx, s, piKey, scopeKey, ei.Key)
				if err != nil {
					return nil, err
				}
				intents = append(intents, termIntents...)
				// Terminate the subprocess itself
				intents = append(intents, &intent.TerminateElementIntent{
					Header: intent.Header{
						Origin:             intent.Internal,
						ProcessInstanceKey: piKey,
					},
					ElementInstanceKey: scopeKey,
				})
				// Activate the boundary error event
				intents = append(intents, &intent.ActivateElementIntent{
					Header: intent.Header{
						Origin:             intent.Internal,
						ProcessInstanceKey: piKey,
					},
					ProcessDefinitionKey: ei.ProcessDefinitionKey,
					ElementId:            boundaryId,
					ElementType:          "boundaryEvent",
					FlowScopeKey:         scopeEI.FlowScopeKey,
				})
				return intents, nil
			}
		}
	}

	// Check for error event subprocess in the same scope as the subprocess
	if scopeKey != 0 && scopeKey != piKey {
		scopeEI, err := s.ProcessInstances().GetElementInstance(ctx, scopeKey)
		if err != nil {
			return nil, err
		}
		if scopeEI != nil {
			espResult := findErrorEventSubProcess(bmi, scopeEI.ElementId, errorCode)
			if espResult != nil {
				return activateErrorEventSubProcess(ctx, s, bmi, espResult, scopeKey, piKey, ei, scopeEI)
			}
		}
	}

	// Check for error event subprocess at process level
	espResult := findErrorEventSubProcessAtProcessLevel(bmi, errorCode)
	if espResult != nil {
		return activateErrorEventSubProcessAtProcess(ctx, s, bmi, espResult, piKey, ei)
	}

	// No boundary error event found — create incident
	return []intent.Intent{
		&intent.CreateIncidentIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ElementInstanceKey: ei.Key,
			ErrorType:          "UNHANDLED_BPMN_ERROR",
			ErrorMessage:       fmt.Sprintf("unhandled error end event %q, error code: %s", ei.ElementId, errorCode),
		},
	}, nil
}

// handleBoundaryEventInterruption checks if the boundary event is interrupting (cancelActivity=true).
// If so, terminates the attached element instance.
func handleBoundaryEventInterruption(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance, piKey uint64) []intent.Intent {
	boundaryEvents := bpmn_model.GetTypedElements[bpmn_model.BoundaryEvent](bmi.ModelInstance)
	for _, be := range boundaryEvents {
		if be.GetId() != ei.ElementId {
			continue
		}

		// cancelActivity defaults to true if not set
		if !be.IsCancelActivity() {
			return nil // non-interrupting — don't terminate attached element
		}

		// Find the attached element instance
		attachedRef := be.GetAttachedToRef()
		if attachedRef == "" {
			return nil
		}

		allEIs, err := s.ProcessInstances().FindElementInstancesByProcessInstance(ctx, piKey)
		if err != nil {
			return nil
		}

		for _, attached := range allEIs {
			if attached.ElementId == attachedRef && attached.State == storage.ElementInstanceActivated {
				return []intent.Intent{
					&intent.TerminateElementIntent{
						Header: intent.Header{
							Origin:             intent.Internal,
							ProcessInstanceKey: piKey,
						},
						ElementInstanceKey: attached.Key,
					},
				}
			}
		}
		break
	}
	return nil
}

// handleLinkThrowEvent checks if the intermediate throw event has a linkEventDefinition.
// If so, finds the matching catch link event with the same name and activates it.
func handleLinkThrowEvent(bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance, piKey uint64) []intent.Intent {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"

	// Find link name on this throw event
	var linkName string
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() != ei.ElementId {
			continue
		}
		dom := fn.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			return nil
		}
		linkDefs := dom.GetChildElementsByNS(bpmnNS, "linkEventDefinition")
		if len(linkDefs) == 0 {
			return nil
		}
		linkName = linkDefs[0].GetAttribute("name")
		break
	}

	if linkName == "" {
		return nil
	}

	// Find catch link event with the same name
	for _, fn := range flowNodes {
		if _, ok := fn.(bpmn_model.IntermediateCatchEvent); !ok {
			continue
		}
		dom := fn.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			continue
		}
		linkDefs := dom.GetChildElementsByNS(bpmnNS, "linkEventDefinition")
		if len(linkDefs) == 0 {
			continue
		}
		catchLinkName := linkDefs[0].GetAttribute("name")
		if catchLinkName == linkName {
			// Found matching catch — activate it (auto-complete to follow outgoing flows)
			return []intent.Intent{
				&intent.ActivateElementIntent{
					Header: intent.Header{
						Origin:             intent.Internal,
						ProcessInstanceKey: piKey,
					},
					ProcessDefinitionKey: ei.ProcessDefinitionKey,
					ElementId:            fn.GetId(),
					ElementType:          "intermediateCatchEvent",
					FlowScopeKey:         ei.FlowScopeKey,
				},
			}
		}
	}

	return nil
}

// handleSignalEndEvent checks if the end event has a signalEventDefinition.
// If so, returns a ThrowSignalIntent.
func handleSignalEndEvent(bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance) []intent.Intent {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"

	endEvents := bpmn_model.GetTypedElements[bpmn_model.EndEvent](bmi.ModelInstance)
	for _, ee := range endEvents {
		if ee.GetId() != ei.ElementId {
			continue
		}
		dom := ee.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			return nil
		}
		sigDefs := dom.GetChildElementsByNS(bpmnNS, "signalEventDefinition")
		if len(sigDefs) == 0 {
			return nil
		}
		sigRef := sigDefs[0].GetAttribute("signalRef")
		signals := bpmn_model.GetTypedElements[bpmn_model.Signal](bmi.ModelInstance)
		sigName := sigRef
		for _, sig := range signals {
			if sig.GetId() == sigRef {
				sigName = sig.GetName()
				break
			}
		}
		return []intent.Intent{
			&intent.ThrowSignalIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: ei.ProcessInstanceKey,
				},
				SignalName: sigName,
			},
		}
	}
	return nil
}

// isTerminateEndEvent checks if the given end event has a terminateEventDefinition child.
func isTerminateEndEvent(bmi *bpmn_model.BpmnModelInstance, elementId string) bool {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"

	endEvents := bpmn_model.GetTypedElements[bpmn_model.EndEvent](bmi.ModelInstance)
	for _, ee := range endEvents {
		if ee.GetId() != elementId {
			continue
		}
		dom := ee.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			return false
		}
		termDefs := dom.GetChildElementsByNS(bpmnNS, "terminateEventDefinition")
		return len(termDefs) > 0
	}
	return false
}

// terminateScopeElements finds all activated element instances in the given scope
// (except the excluded one) and emits TerminateElementIntent for each.
func terminateScopeElements(ctx context.Context, s storage.Store, piKey, scopeKey, excludeKey uint64) ([]intent.Intent, error) {
	allEIs, err := s.ProcessInstances().FindElementInstancesByProcessInstance(ctx, piKey)
	if err != nil {
		return nil, err
	}

	var intents []intent.Intent
	for _, ei := range allEIs {
		if ei.Key == excludeKey {
			continue
		}
		if ei.FlowScopeKey == scopeKey && ei.State == storage.ElementInstanceActivated {
			intents = append(intents, &intent.TerminateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: piKey,
				},
				ElementInstanceKey: ei.Key,
			})
		}
	}
	return intents, nil
}

// cancelEventBasedGatewaySiblings checks if the completed catch event was downstream
// of an event-based gateway. If so, terminates all sibling catch event instances.
func cancelEventBasedGatewaySiblings(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, completedEI *storage.ElementInstance, piKey uint64) []intent.Intent {
	// Find the completed element in BPMN model
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	var completedNode bpmn_model.FlowNode
	for _, fn := range flowNodes {
		if fn.GetId() == completedEI.ElementId {
			completedNode = fn
			break
		}
	}
	if completedNode == nil {
		return nil
	}

	// Check if any incoming flow comes from an event-based gateway
	var ebGateway bpmn_model.FlowNode
	for _, sf := range completedNode.GetIncomingSequenceFlows() {
		src := sf.GetSource()
		if src == nil {
			continue
		}
		if _, ok := src.(bpmn_model.EventBasedGateway); ok {
			ebGateway = src
			break
		}
	}
	if ebGateway == nil {
		return nil
	}

	// Find all sibling catch event element IDs (other outgoing targets of the gateway)
	siblingIds := map[string]bool{}
	for _, sf := range ebGateway.GetOutgoingSequenceFlows() {
		target := sf.GetTarget()
		if target != nil && target.GetId() != completedEI.ElementId {
			siblingIds[target.GetId()] = true
		}
	}

	if len(siblingIds) == 0 {
		return nil
	}

	// Find active sibling element instances and terminate them
	allEIs, err := s.ProcessInstances().FindElementInstancesByProcessInstance(ctx, piKey)
	if err != nil {
		return nil
	}

	var intents []intent.Intent
	for _, ei := range allEIs {
		if siblingIds[ei.ElementId] && ei.State == storage.ElementInstanceActivated {
			intents = append(intents, &intent.TerminateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: piKey,
				},
				ElementInstanceKey: ei.Key,
			})
		}
	}

	return intents
}
