package gateway

import (
	"context"
	"time"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ActivateJobs(ctx context.Context, req *pb.ActivateJobsRequest) (*pb.ActivateJobsResponse, error) {
	if req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}
	if req.Worker == "" {
		return nil, status.Error(codes.InvalidArgument, "worker is required")
	}

	maxJobs := int(req.MaxJobs)
	if maxJobs <= 0 {
		maxJobs = 1
	}

	timeout := 5 * time.Minute
	if req.Timeout != nil {
		timeout = req.Timeout.AsDuration()
	}

	// Find activatable jobs from storage
	jobs, err := s.store.Jobs().FindActivatable(ctx, req.Type, maxJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find jobs: %v", err)
	}

	resp := &pb.ActivateJobsResponse{}

	for _, job := range jobs {
		// Submit activate intent for each job
		activateIntent := &intent.ActivateJobIntent{
			Header: intent.Header{
				Origin:             intent.External,
				ProcessInstanceKey: job.ProcessInstanceKey,
			},
			JobKey:  job.Key,
			Worker:  req.Worker,
			Timeout: timeout,
		}
		s.processor.Submit(activateIntent)

		// Find element id for the job
		var elementId string
		ei, err := s.store.ProcessInstances().GetElementInstance(ctx, job.ElementInstanceKey)
		if err == nil && ei != nil {
			elementId = ei.ElementId
		}

		var variables string
		if len(job.Variables) > 0 {
			variables = string(job.Variables)
		}

		resp.Jobs = append(resp.Jobs, &pb.ActivatedJob{
			Key:                  job.Key,
			ProcessInstanceKey:   job.ProcessInstanceKey,
			ProcessDefinitionKey: job.ProcessDefinitionKey,
			ElementId:            elementId,
			Type:                 job.Type,
			Retries:              int32(job.Retries),
			Variables:            variables,
		})
	}

	return resp, nil
}
