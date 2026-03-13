package behavior

import (
	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// handleCompensationThrow handles compensation throw (intermediate throw or end event).
// Finds boundary compensation events in the scope and activates their linked handlers.
func handleCompensationThrow(bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance, piKey uint64) []intent.Intent {
	// Verify this element has a compensateEventDefinition
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	hasCompensation := false
	for _, fn := range flowNodes {
		if fn.GetId() != ei.ElementId {
			continue
		}
		dom := fn.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			break
		}
		compDefs := dom.GetChildElementsByNS(bpmnNS, "compensateEventDefinition")
		if len(compDefs) > 0 {
			hasCompensation = true
		}
		break
	}

	if !hasCompensation {
		return nil
	}

	// Find all boundary compensation events and their associated handler activities
	// BoundaryEvent with compensateEventDefinition → Association → handler activity
	boundaryEvents := bpmn_model.GetTypedElements[bpmn_model.BoundaryEvent](bmi.ModelInstance)
	associations := bpmn_model.GetTypedElements[bpmn_model.Association](bmi.ModelInstance)

	var intents []intent.Intent
	for _, be := range boundaryEvents {
		dom := be.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			continue
		}
		compDefs := dom.GetChildElementsByNS(bpmnNS, "compensateEventDefinition")
		if len(compDefs) == 0 {
			continue
		}

		// Find association from this boundary event to a handler activity
		handlerId := findCompensationHandler(associations, be.GetId())
		if handlerId == "" {
			continue
		}

		// Determine handler element type
		handlerType := "unknown"
		for _, fn := range flowNodes {
			if fn.GetId() == handlerId {
				handlerType = resolveElementType(fn)
				break
			}
		}

		intents = append(intents, &intent.ActivateElementIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ProcessDefinitionKey: ei.ProcessDefinitionKey,
			ElementId:            handlerId,
			ElementType:          handlerType,
			FlowScopeKey:         ei.FlowScopeKey,
		})
	}

	return intents
}

// findCompensationHandler finds the target activity id linked by Association from a boundary event.
func findCompensationHandler(associations []bpmn_model.Association, boundaryEventId string) string {
	for _, assoc := range associations {
		if assoc.GetSourceRef() == boundaryEventId {
			return assoc.GetTargetRef()
		}
	}
	return ""
}
