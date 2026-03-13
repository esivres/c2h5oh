package gateway

import (
	"context"

	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/intent"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) BroadcastSignal(ctx context.Context, req *pb.BroadcastSignalRequest) (*pb.BroadcastSignalResponse, error) {
	if req.SignalName == "" {
		return nil, status.Error(codes.InvalidArgument, "signal_name is required")
	}

	var variables []byte
	if req.Variables != "" {
		variables = []byte(req.Variables)
	}

	signalIntent := &intent.ThrowSignalIntent{
		Header:     intent.Header{Origin: intent.External},
		SignalName: req.SignalName,
		Variables:  variables,
	}
	s.processor.Submit(signalIntent)

	return &pb.BroadcastSignalResponse{
		Key: signalIntent.Key,
	}, nil
}
