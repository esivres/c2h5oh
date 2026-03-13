package gateway

import (
	"context"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteResource(ctx context.Context, req *pb.DeleteResourceRequest) (*pb.DeleteResourceResponse, error) {
	if req.ResourceKey == 0 {
		return nil, status.Error(codes.InvalidArgument, "resource_key is required")
	}

	deleteIntent := &intent.DeleteResourceIntent{
		Header:      intent.Header{Origin: intent.External},
		ResourceKey: req.ResourceKey,
	}
	s.processor.Submit(deleteIntent)

	return &pb.DeleteResourceResponse{}, nil
}
