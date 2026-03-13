package gateway

import (
	"context"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateJobRetries(ctx context.Context, req *pb.UpdateJobRetriesRequest) (*pb.UpdateJobRetriesResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}
	if req.Retries < 1 {
		return nil, status.Error(codes.InvalidArgument, "retries must be > 0")
	}

	// Verify job exists
	job, err := s.store.Jobs().GetByKey(ctx, req.JobKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	updateIntent := &intent.UpdateJobRetriesIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:  req.JobKey,
		Retries: int(req.Retries),
	}
	s.processor.Submit(updateIntent)

	return &pb.UpdateJobRetriesResponse{}, nil
}
