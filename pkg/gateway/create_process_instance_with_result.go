package gateway

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateProcessInstanceWithResult(ctx context.Context, req *pb.CreateProcessInstanceWithResultRequest) (*pb.CreateProcessInstanceWithResultResponse, error) {
	if req.ProcessId == "" {
		return nil, status.Error(codes.InvalidArgument, "process_id is required")
	}

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

	// Subscribe before submitting to avoid race
	s.processor.Submit(createIntent)
	piKey := createIntent.Key

	notify := s.completionNotifier.Subscribe(piKey)
	defer s.completionNotifier.Unsubscribe(piKey, notify)

	// Determine timeout
	timeout := 10 * time.Second
	if req.RequestTimeout != nil {
		timeout = req.RequestTimeout.AsDuration()
	}

	deadline := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "request canceled")
		case <-deadline:
			return nil, status.Errorf(codes.DeadlineExceeded, "process instance %d did not complete within timeout", piKey)
		case <-notify:
			// Check if the process actually completed
			pi, err := s.store.ProcessInstances().GetInstance(ctx, piKey)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check process state: %v", err)
			}
			if pi == nil {
				continue // spurious notification
			}

			// Collect result variables from the process scope
			resultVars, err := s.collectResultVariables(ctx, piKey, req.FetchVariables)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to collect variables: %v", err)
			}

			return &pb.CreateProcessInstanceWithResultResponse{
				ProcessDefinitionKey: def.Key,
				BpmnProcessId:        req.ProcessId,
				ProcessInstanceKey:   piKey,
				Variables:            resultVars,
			}, nil
		}
	}
}

// collectResultVariables reads variables from the process instance scope
// and returns them as a JSON string. If fetchVariables is non-empty, only
// those names are included.
func (s *Server) collectResultVariables(ctx context.Context, piKey uint64, fetchVariables []string) (string, error) {
	vars, err := s.store.Variables().FindByScope(ctx, piKey)
	if err != nil {
		return "", err
	}

	filter := make(map[string]bool, len(fetchVariables))
	for _, name := range fetchVariables {
		filter[name] = true
	}

	result := make(map[string]json.RawMessage, len(vars))
	for _, v := range vars {
		if len(filter) > 0 && !filter[v.Name] {
			continue
		}
		result[v.Name] = json.RawMessage(v.Value)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
