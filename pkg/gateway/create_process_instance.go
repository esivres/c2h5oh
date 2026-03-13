package gateway

import (
	"context"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateProcessInstance(ctx context.Context, req *pb.CreateProcessInstanceRequest) (*pb.CreateProcessInstanceResponse, error) {
	if req.ProcessId == "" {
		return nil, status.Error(codes.InvalidArgument, "process_id is required")
	}

	// Resolve process definition to validate it exists
	def, err := s.store.ProcessDefinitions().FindLatestByProcessId(ctx, req.ProcessId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find process: %v", err)
	}
	if def == nil {
		return nil, status.Errorf(codes.NotFound, "process %q not found", req.ProcessId)
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	createIntent := &intent.CreateProcessInstanceIntent{
		Header:               intent.Header{Origin: intent.External},
		ProcessDefinitionKey: def.Key,
		BpmnProcessId:        req.ProcessId,
		Variables:            variables,
	}

	// Submit assigns the key; after this call, createIntent.Key is the process instance key
	s.processor.Submit(createIntent)

	return &pb.CreateProcessInstanceResponse{
		ProcessInstanceKey:   createIntent.Key,
		ProcessDefinitionKey: def.Key,
	}, nil
}
