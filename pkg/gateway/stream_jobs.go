package gateway

import (
	"time"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) StreamActivatedJobs(req *pb.StreamActivatedJobsRequest, stream grpc.ServerStreamingServer[pb.ActivatedJob]) error {
	if req.Type == "" {
		return status.Error(codes.InvalidArgument, "type is required")
	}
	if req.Worker == "" {
		return status.Error(codes.InvalidArgument, "worker is required")
	}

	timeout := 5 * time.Minute
	if req.Timeout != nil {
		timeout = req.Timeout.AsDuration()
	}

	maxJobs := int(req.MaxJobs)
	if maxJobs <= 0 {
		maxJobs = 1
	}

	ctx := stream.Context()

	// Subscribe to notifications for this job type
	notify := s.jobNotifier.Subscribe(req.Type)
	defer s.jobNotifier.Unsubscribe(req.Type, notify)

	// Initial poll for already-existing jobs
	if err := s.sendActivatableJobs(stream, req.Type, req.Worker, timeout, maxJobs); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-notify:
			if err := s.sendActivatableJobs(stream, req.Type, req.Worker, timeout, maxJobs); err != nil {
				return err
			}
		}
	}
}

// sendActivatableJobs finds activatable jobs, activates them, and sends to the stream.
func (s *Server) sendActivatableJobs(stream grpc.ServerStreamingServer[pb.ActivatedJob], jobType, worker string, timeout time.Duration, maxJobs int) error {
	ctx := stream.Context()

	jobs, err := s.store.Jobs().FindActivatable(ctx, jobType, maxJobs)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to find jobs: %v", err)
	}

	for _, job := range jobs {
		activateIntent := &intent.ActivateJobIntent{
			Header: intent.Header{
				Origin:             intent.External,
				ProcessInstanceKey: job.ProcessInstanceKey,
			},
			JobKey:  job.Key,
			Worker:  worker,
			Timeout: timeout,
		}
		s.processor.Submit(activateIntent)

		var elementId string
		ei, err := s.store.ProcessInstances().GetElementInstance(ctx, job.ElementInstanceKey)
		if err == nil && ei != nil {
			elementId = ei.ElementId
		}

		var variables string
		if len(job.Variables) > 0 {
			variables = string(job.Variables)
		}

		if err := stream.Send(&pb.ActivatedJob{
			Key:                  job.Key,
			ProcessInstanceKey:   job.ProcessInstanceKey,
			ProcessDefinitionKey: job.ProcessDefinitionKey,
			ElementId:            elementId,
			Type:                 job.Type,
			Retries:              int32(job.Retries),
			Variables:            variables,
		}); err != nil {
			return err
		}
	}

	return nil
}
