package gateway

import (
	"context"
	"fmt"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ActivateElement(ctx context.Context, req *pb.ActivateElementRequest) (*pb.ActivateElementResponse, error) {
	if req.ProcessInstanceKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "process_instance_key is required")
	}
	if req.ScopeKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "scope_key is required")
	}
	if req.ElementId == "" {
		return nil, status.Error(codes.InvalidArgument, "element_id is required")
	}

	// Validate scope_key is an ad-hoc subprocess element instance
	scopeEI, err := s.store.ProcessInstances().GetElementInstance(ctx, req.ScopeKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find element instance: %v", err)
	}
	if scopeEI == nil {
		return nil, status.Errorf(codes.NotFound, "element instance %d not found", req.ScopeKey)
	}
	if scopeEI.ElementType != "adHocSubProcess" {
		return nil, status.Errorf(codes.FailedPrecondition, "element instance %d is not an ad-hoc subprocess (type=%s)", req.ScopeKey, scopeEI.ElementType)
	}

	// Load process definition and parse BPMN
	def, err := s.store.ProcessDefinitions().FindByKey(ctx, scopeEI.ProcessDefinitionKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find process definition: %v", err)
	}
	if def == nil {
		return nil, status.Errorf(codes.NotFound, "process definition %d not found", scopeEI.ProcessDefinitionKey)
	}

	bmi, err := bpmn_model.ReadFromBytes(def.Content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse BPMN: %v", err)
	}

	// Find the ad-hoc subprocess and validate element_id exists inside it
	elementType, jobType, err := resolveAdHocChildElement(bmi, scopeEI.ElementId, req.ElementId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// If variables provided, set them on the ad-hoc scope
	if req.Variables != "" {
		setVarsIntent := &intent.SetVariablesIntent{
			Header: intent.Header{
				Origin:             intent.External,
				ProcessInstanceKey: req.ProcessInstanceKey,
			},
			ScopeKey:  req.ScopeKey,
			Variables: []byte(req.Variables),
		}
		s.processor.Submit(setVarsIntent)
	}

	// Submit ActivateElementIntent
	activateIntent := &intent.ActivateElementIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: req.ProcessInstanceKey,
		},
		ProcessDefinitionKey: scopeEI.ProcessDefinitionKey,
		ElementId:            req.ElementId,
		ElementType:          elementType,
		FlowScopeKey:         req.ScopeKey,
		JobType:              jobType,
	}
	s.processor.Submit(activateIntent)

	return &pb.ActivateElementResponse{
		ElementInstanceKey: activateIntent.Key,
	}, nil
}

// resolveAdHocChildElement finds the element with the given elementId inside the ad-hoc subprocess
// identified by adHocId. Returns the element type and job type (from ZeebeTaskDefinition).
func resolveAdHocChildElement(bmi *bpmn_model.BpmnModelInstance, adHocId, elementId string) (string, string, error) {
	adHocs := bpmn_model.GetTypedElements[bpmn_model.AdHocSubProcess](bmi.ModelInstance)
	for _, ah := range adHocs {
		if ah.GetId() != adHocId {
			continue
		}

		for _, fe := range ah.GetFlowElements() {
			fn, ok := fe.(bpmn_model.FlowNode)
			if !ok {
				continue
			}
			if fn.GetId() != elementId {
				continue
			}

			elementType := resolveElementTypeFromNode(fn)
			jobType := resolveJobTypeFromNode(bmi, fn)
			return elementType, jobType, nil
		}

		return "", "", fmt.Errorf("element %q not found inside ad-hoc subprocess %q", elementId, adHocId)
	}

	return "", "", fmt.Errorf("ad-hoc subprocess %q not found in process definition", adHocId)
}

// resolveElementTypeFromNode returns the BPMN element type string for a flow node.
func resolveElementTypeFromNode(node bpmn_model.FlowNode) string {
	switch node.(type) {
	case bpmn_model.StartEvent:
		return "startEvent"
	case bpmn_model.EndEvent:
		return "endEvent"
	case bpmn_model.ServiceTask:
		return "serviceTask"
	case bpmn_model.UserTask:
		return "userTask"
	case bpmn_model.ScriptTask:
		return "scriptTask"
	case bpmn_model.BusinessRuleTask:
		return "businessRuleTask"
	case bpmn_model.SendTask:
		return "sendTask"
	case bpmn_model.ReceiveTask:
		return "receiveTask"
	case bpmn_model.ManualTask:
		return "manualTask"
	case bpmn_model.ExclusiveGateway:
		return "exclusiveGateway"
	case bpmn_model.ParallelGateway:
		return "parallelGateway"
	case bpmn_model.InclusiveGateway:
		return "inclusiveGateway"
	case bpmn_model.EventBasedGateway:
		return "eventBasedGateway"
	case bpmn_model.AdHocSubProcess:
		return "adHocSubProcess"
	case bpmn_model.SubProcess:
		return "subProcess"
	case bpmn_model.CallActivity:
		return "callActivity"
	case bpmn_model.IntermediateCatchEvent:
		return "intermediateCatchEvent"
	case bpmn_model.IntermediateThrowEvent:
		return "intermediateThrowEvent"
	case bpmn_model.BoundaryEvent:
		return "boundaryEvent"
	default:
		return "unknown"
	}
}

// resolveJobTypeFromNode extracts the job type from a ZeebeTaskDefinition extension element.
func resolveJobTypeFromNode(_ *bpmn_model.BpmnModelInstance, node bpmn_model.FlowNode) string {
	if be, ok := node.(bpmn_model.BaseElement); ok {
		td, found := bpmn_model.GetSingleExtensionElement[bpmn_model.ZeebeTaskDefinition](be)
		if found {
			return td.GetType()
		}
	}
	return ""
}
