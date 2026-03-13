package gateway

import (
	"context"
	"crypto/sha256"
	"time"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/bpmn_model"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeployProcess(ctx context.Context, req *pb.DeployProcessRequest) (*pb.DeployProcessResponse, error) {
	if len(req.Definition) == 0 {
		return nil, status.Error(codes.InvalidArgument, "definition is required")
	}

	// Parse BPMN to extract process id
	bmi, err := bpmn_model.ReadFromBytes(req.Definition)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid BPMN XML: %v", err)
	}

	processes := bpmn_model.GetTypedElements[bpmn_model.Process](bmi.ModelInstance)
	if len(processes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "BPMN XML contains no process")
	}

	processId := processes[0].GetId()
	if processId == "" {
		return nil, status.Error(codes.InvalidArgument, "process has no id")
	}

	contentHash := sha256.Sum256(req.Definition)

	// Idempotent: if same content already deployed, return existing definition
	existing, err := s.store.ProcessDefinitions().FindByContentHash(ctx, processId, contentHash[:])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check existing deployment: %v", err)
	}
	if existing != nil {
		return &pb.DeployProcessResponse{
			Key:       existing.Key,
			ProcessId: processId,
			Version:   existing.Version,
		}, nil
	}

	name := req.Name
	if name == "" {
		name = processId
	}

	// Submit deploy intent
	deployIntent := &intent.DeployProcessIntent{
		Header:        intent.Header{Origin: intent.External},
		BpmnProcessId: processId,
		Name:          name,
		Content:       req.Definition,
		ContentHash:   contentHash[:],
	}
	notify := s.deployNotifier.Subscribe(processId)
	defer s.deployNotifier.Unsubscribe(processId, notify)

	s.processor.Submit(deployIntent)

	// Wait for the deploy notification or timeout
	const deployTimeout = 10 * time.Second
	deadline := time.After(deployTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "request canceled")
		case <-deadline:
			return nil, status.Error(codes.DeadlineExceeded, "deploy timed out")
		case <-notify:
			def, err := s.store.ProcessDefinitions().FindByContentHash(ctx, processId, contentHash[:])
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check deployment: %v", err)
			}
			if def != nil {
				return &pb.DeployProcessResponse{
					Key:       def.Key,
					ProcessId: processId,
					Version:   def.Version,
				}, nil
			}
		}
	}
}
