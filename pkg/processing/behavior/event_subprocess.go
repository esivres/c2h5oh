package behavior

import (
	"context"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// errorEventSubProcessMatch holds info about a matching error event subprocess.
type errorEventSubProcessMatch struct {
	SubProcessId string // the event subprocess element id
	StartEventId string // the error start event id inside
	Interrupting bool   // whether the start event is interrupting
}

// findErrorEventSubProcess searches for an event subprocess with an error start event
// inside the given parent subprocess element.
func findErrorEventSubProcess(bmi *bpmn_model.BpmnModelInstance, parentElementId string, errorCode string) *errorEventSubProcessMatch {
	subProcesses := bpmn_model.GetTypedElements[bpmn_model.SubProcess](bmi.ModelInstance)
	for _, sp := range subProcesses {
		if sp.GetId() != parentElementId {
			continue
		}

		// Look for event subprocesses among flow elements
		for _, fe := range sp.GetFlowElements() {
			esp, ok := fe.(bpmn_model.SubProcess)
			if !ok || !esp.IsTriggeredByEvent() {
				continue
			}

			if match := matchErrorStartEvent(bmi, esp, errorCode); match != nil {
				return match
			}
		}
		break
	}
	return nil
}

// findErrorEventSubProcessAtProcessLevel searches for an event subprocess with an error start event
// at the process level (not inside another subprocess).
func findErrorEventSubProcessAtProcessLevel(bmi *bpmn_model.BpmnModelInstance, errorCode string) *errorEventSubProcessMatch {
	processes := bpmn_model.GetTypedElements[bpmn_model.Process](bmi.ModelInstance)
	for _, p := range processes {
		for _, fe := range p.GetFlowElements() {
			esp, ok := fe.(bpmn_model.SubProcess)
			if !ok || !esp.IsTriggeredByEvent() {
				continue
			}

			if match := matchErrorStartEvent(bmi, esp, errorCode); match != nil {
				return match
			}
		}
	}
	return nil
}

// matchErrorStartEvent checks if the event subprocess has an error start event matching the error code.
func matchErrorStartEvent(
	bmi *bpmn_model.BpmnModelInstance, esp bpmn_model.SubProcess,
	errorCode string,
) *errorEventSubProcessMatch {
	for _, inner := range esp.GetFlowElements() {
		se, ok := inner.(bpmn_model.StartEvent)
		if !ok {
			continue
		}

		dom := se.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			continue
		}

		errorDefs := dom.GetChildElementsByNS(bpmnNS, "errorEventDefinition")
		if len(errorDefs) == 0 {
			continue
		}

		// Check error code match
		errorRef := errorDefs[0].GetAttribute("errorRef")
		if errorRef != "" {
			if !matchesErrorCode(bmi, errorRef, errorCode) {
				continue
			}
		}
		// errorRef="" means catch-all

		// Check interrupting: start event in event subprocess is interrupting by default.
		// Non-interrupting has isInterrupting="false" on the start event DOM.
		interrupting := true
		if isInterruptingAttr := dom.GetAttribute("isInterrupting"); isInterruptingAttr == "false" {
			interrupting = false
		}

		return &errorEventSubProcessMatch{
			SubProcessId: esp.GetId(),
			StartEventId: se.GetId(),
			Interrupting: interrupting,
		}
	}
	return nil
}

// activateErrorEventSubProcess activates an error event subprocess found inside a parent subprocess scope.
func activateErrorEventSubProcess(
	ctx context.Context, s storage.Store, _ *bpmn_model.BpmnModelInstance,
	match *errorEventSubProcessMatch, scopeKey, piKey uint64,
	errorEI *storage.ElementInstance, _ *storage.ElementInstance,
) ([]intent.Intent, error) {
	var intents []intent.Intent

	if match.Interrupting {
		// Terminate all active elements in the scope
		termIntents, err := terminateScopeElements(ctx, s, piKey, scopeKey, errorEI.Key)
		if err != nil {
			return nil, err
		}
		intents = append(intents, termIntents...)
	}

	// Activate the event subprocess
	intents = append(intents, &intent.ActivateElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: piKey,
		},
		ProcessDefinitionKey: errorEI.ProcessDefinitionKey,
		ElementId:            match.SubProcessId,
		ElementType:          "subProcess",
		FlowScopeKey:         scopeKey,
	})

	return intents, nil
}

// activateErrorEventSubProcessAtProcess activates an error event subprocess at process level.
func activateErrorEventSubProcessAtProcess(
	ctx context.Context, s storage.Store, _ *bpmn_model.BpmnModelInstance,
	match *errorEventSubProcessMatch, piKey uint64,
	errorEI *storage.ElementInstance,
) ([]intent.Intent, error) {
	var intents []intent.Intent

	if match.Interrupting {
		termIntents, err := terminateScopeElements(ctx, s, piKey, piKey, errorEI.Key)
		if err != nil {
			return nil, err
		}
		intents = append(intents, termIntents...)
	}

	intents = append(intents, &intent.ActivateElementIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: piKey,
		},
		ProcessDefinitionKey: errorEI.ProcessDefinitionKey,
		ElementId:            match.SubProcessId,
		ElementType:          "subProcess",
		FlowScopeKey:         piKey,
	})

	return intents, nil
}
