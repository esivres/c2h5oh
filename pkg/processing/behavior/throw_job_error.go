package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// throwJobError handles ThrowJobErrorIntent.
//
// 1. Mark job as ErrorThrown.
// 2. Look for a boundary error event on the element that owns the job.
// 3. If found → TerminateElement (current) + ActivateElement (boundary event).
// 4. If not found → CreateIncident.
func throwJobError(ctx context.Context, s storage.Store, i *intent.ThrowJobErrorIntent) ([]intent.Intent, error) {
	job, err := s.Jobs().GetByKey(ctx, i.JobKey)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found: %d", i.JobKey)
	}

	if err := s.Jobs().ThrowError(ctx, i.JobKey, i.ErrorCode, i.ErrorMessage); err != nil {
		return nil, err
	}

	// Get the element instance to find the element id
	ei, err := s.ProcessInstances().GetElementInstance(ctx, job.ElementInstanceKey)
	if err != nil {
		return nil, err
	}
	if ei == nil {
		return nil, fmt.Errorf("element instance not found: %d", job.ElementInstanceKey)
	}

	// Load BPMN model to find boundary error events
	def, err := s.ProcessDefinitions().FindByKey(ctx, job.ProcessDefinitionKey)
	if err != nil {
		return nil, err
	}
	if def != nil {
		bmi, err := bpmn_model.ReadFromBytes(def.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse BPMN: %w", err)
		}

		boundaryId := findBoundaryErrorEvent(bmi, ei.ElementId, i.ErrorCode)
		if boundaryId != "" {
			var intents []intent.Intent

			// Terminate the current element (cancel activity)
			intents = append(intents, &intent.TerminateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ElementInstanceKey: job.ElementInstanceKey,
			})

			// Activate the boundary error event
			intents = append(intents, &intent.ActivateElementIntent{
				Header: intent.Header{
					Origin:             intent.Internal,
					ProcessInstanceKey: i.ProcessInstanceKey,
				},
				ProcessDefinitionKey: job.ProcessDefinitionKey,
				ElementId:            boundaryId,
				ElementType:          "boundaryEvent",
				FlowScopeKey:         ei.FlowScopeKey,
			})

			return intents, nil
		}
	}

	// No boundary error event found → create incident
	return []intent.Intent{
		&intent.CreateIncidentIntent{
			Header: intent.Header{
				Origin:             intent.Internal,
				ProcessInstanceKey: i.ProcessInstanceKey,
			},
			ElementInstanceKey: job.ElementInstanceKey,
			JobKey:             i.JobKey,
			ErrorType:          "UNHANDLED_BPMN_ERROR",
			ErrorMessage:       fmt.Sprintf("error code: %s, message: %s", i.ErrorCode, i.ErrorMessage),
		},
	}, nil
}

// findBoundaryErrorEvent looks for a boundary error event attached to the given element.
// If errorCode is non-empty, it matches boundary events with a matching error reference.
// A boundary event with no error reference (catch-all) matches any error code.
// Returns the boundary event id, or empty string if not found.
func findBoundaryErrorEvent(bmi *bpmn_model.BpmnModelInstance, elementId string, errorCode string) string {
	boundaryEvents := bpmn_model.GetTypedElements[bpmn_model.BoundaryEvent](bmi.ModelInstance)

	var catchAll string

	for _, be := range boundaryEvents {
		if be.GetAttachedToRef() != elementId {
			continue
		}

		// Check if this boundary event has an ErrorEventDefinition
		// We look for ErrorEventDefinition children under this boundary event's DOM element
		errorDefs := findErrorEventDefinitions(bmi, be.GetId())
		if len(errorDefs) == 0 {
			continue
		}

		for _, errDef := range errorDefs {
			errRef := errDef.GetErrorRef()
			if errRef == "" {
				// Catch-all error boundary event (no specific error ref)
				catchAll = be.GetId()
				continue
			}

			// Match error code via the Error element
			if matchesErrorCode(bmi, errRef, errorCode) {
				return be.GetId()
			}
		}
	}

	return catchAll
}

// findErrorEventDefinitions finds ErrorEventDefinition elements that are children
// of the boundary event with the given id.
func findErrorEventDefinitions(bmi *bpmn_model.BpmnModelInstance, boundaryEventId string) []bpmn_model.ErrorEventDefinition {
	// Get all ErrorEventDefinitions in the model and filter by parent
	allErrorDefs := bpmn_model.GetTypedElements[bpmn_model.ErrorEventDefinition](bmi.ModelInstance)

	var result []bpmn_model.ErrorEventDefinition
	for _, ed := range allErrorDefs {
		// Check if this error definition's parent element is our boundary event
		if ed.GetDomElement() != nil && ed.GetDomElement().Parent() != nil {
			parentId := ed.GetDomElement().Parent().GetAttribute("id")
			if parentId == boundaryEventId {
				result = append(result, ed)
			}
		}
	}
	return result
}

// matchesErrorCode checks if a BPMN Error element (referenced by errorRef)
// has the given error code.
func matchesErrorCode(bmi *bpmn_model.BpmnModelInstance, errorRef string, errorCode string) bool {
	if errorCode == "" {
		return true // no specific code to match
	}

	errors := bpmn_model.GetTypedElements[bpmn_model.Error](bmi.ModelInstance)
	for _, e := range errors {
		if e.GetId() == errorRef {
			return e.GetErrorCode() == errorCode
		}
	}

	// If error element not found, treat errorRef as the error code itself
	return errorRef == errorCode
}
