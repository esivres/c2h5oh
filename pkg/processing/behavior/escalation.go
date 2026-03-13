package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// handleEscalationThrow handles escalation throw events (intermediate throw or end event).
// Finds a matching boundary escalation event on the parent scope or creates an incident.
func handleEscalationThrow(ctx context.Context, s storage.Store, bmi *bpmn_model.BpmnModelInstance, ei *storage.ElementInstance, piKey uint64) ([]intent.Intent, error) {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"
	escalationCode := ""

	// Find the element in BPMN and extract escalation code
	flowNodes := bpmn_model.GetTypedElements[bpmn_model.FlowNode](bmi.ModelInstance)
	for _, fn := range flowNodes {
		if fn.GetId() != ei.ElementId {
			continue
		}
		dom := fn.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			break
		}
		escDefs := dom.GetChildElementsByNS(bpmnNS, "escalationEventDefinition")
		if len(escDefs) == 0 {
			return nil, nil // not an escalation event
		}
		escRef := escDefs[0].GetAttribute("escalationRef")
		if escRef != "" {
			escalations := bpmn_model.GetTypedElements[bpmn_model.Escalation](bmi.ModelInstance)
			for _, e := range escalations {
				if e.GetId() == escRef {
					escalationCode = e.GetEscalationCode()
					break
				}
			}
			if escalationCode == "" {
				escalationCode = escRef
			}
		}
		break
	}

	// Walk up scope chain to find boundary escalation event
	scopeKey := ei.FlowScopeKey
	if scopeKey != 0 && scopeKey != piKey {
		scopeEI, err := s.ProcessInstances().GetElementInstance(ctx, scopeKey)
		if err != nil {
			return nil, err
		}
		if scopeEI != nil {
			boundaryId := findBoundaryEscalationEvent(bmi, scopeEI.ElementId, escalationCode)
			if boundaryId != "" {
				return []intent.Intent{
					&intent.ActivateElementIntent{
						Header: intent.Header{
							Origin:             intent.Internal,
							ProcessInstanceKey: piKey,
						},
						ProcessDefinitionKey: ei.ProcessDefinitionKey,
						ElementId:            boundaryId,
						ElementType:          "boundaryEvent",
						FlowScopeKey:         scopeEI.FlowScopeKey,
					},
				}, nil
			}
		}
	}

	// No handler found — create incident
	return []intent.Intent{
		&intent.CreateIncidentIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: piKey,
			},
			ElementInstanceKey: ei.Key,
			ErrorType:          "UNHANDLED_ESCALATION",
			ErrorMessage:       fmt.Sprintf("unhandled escalation %q, code: %s", ei.ElementId, escalationCode),
		},
	}, nil
}

// findBoundaryEscalationEvent looks for a boundary escalation event attached to the given element.
func findBoundaryEscalationEvent(bmi *bpmn_model.BpmnModelInstance, elementId string, escalationCode string) string {
	bpmnNS := "http://www.omg.org/spec/BPMN/20100524/MODEL"
	boundaryEvents := bpmn_model.GetTypedElements[bpmn_model.BoundaryEvent](bmi.ModelInstance)

	var catchAll string

	for _, be := range boundaryEvents {
		if be.GetAttachedToRef() != elementId {
			continue
		}

		dom := be.(bpmn_model.BaseElement).GetDomElement()
		if dom == nil {
			continue
		}

		escDefs := dom.GetChildElementsByNS(bpmnNS, "escalationEventDefinition")
		if len(escDefs) == 0 {
			continue
		}

		escRef := escDefs[0].GetAttribute("escalationRef")
		if escRef == "" {
			// Catch-all escalation
			catchAll = be.GetId()
			continue
		}

		// Match escalation code
		if matchesEscalationCode(bmi, escRef, escalationCode) {
			return be.GetId()
		}
	}

	return catchAll
}

func matchesEscalationCode(bmi *bpmn_model.BpmnModelInstance, escalationRef string, escalationCode string) bool {
	if escalationCode == "" {
		return true
	}
	escalations := bpmn_model.GetTypedElements[bpmn_model.Escalation](bmi.ModelInstance)
	for _, e := range escalations {
		if e.GetId() == escalationRef {
			return e.GetEscalationCode() == escalationCode
		}
	}
	return escalationRef == escalationCode
}
