package gateway

import (
	"context"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CompleteJob(ctx context.Context, req *pb.CompleteJobRequest) (*pb.CompleteJobResponse, error) {
	if req.JobKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "job_key is required")
	}

	// Verify job exists
	job, err := s.store.Jobs().GetByKey(ctx, req.JobKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find job: %v", err)
	}
	if job == nil {
		return nil, status.Errorf(codes.NotFound, "job %d not found", req.JobKey)
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	completeIntent := &intent.CompleteJobIntent{
		Header: intent.Header{
			Origin:             intent.External,
			ProcessInstanceKey: job.ProcessInstanceKey,
		},
		JobKey:    req.JobKey,
		Variables: variables,
	}
	s.processor.Submit(completeIntent)

	return &pb.CompleteJobResponse{}, nil
}
