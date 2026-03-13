package gateway

import (
	"context"
	"time"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) PublishMessage(_ context.Context, req *pb.PublishMessageRequest) (*pb.PublishMessageResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	ttl := 5 * time.Minute
	if req.TimeToLive != nil {
		ttl = req.TimeToLive.AsDuration()
	}

	msgIntent := &intent.PublishMessageIntent{
		Header:         intent.Header{Origin: intent.External},
		MessageName:    req.Name,
		CorrelationKey: req.CorrelationKey,
		Variables:      variables,
		TTL:            ttl,
	}
	s.processor.Submit(msgIntent)

	return &pb.PublishMessageResponse{
		Key: msgIntent.Key,
	}, nil
}
