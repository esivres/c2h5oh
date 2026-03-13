// Package gateway implements the gRPC command API for the BPMN engine.
// It translates external gRPC requests into engine intents.
package gateway

import (
	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/processing/engine"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// Server implements the GatewayAPI gRPC service.
type Server struct {
	pb.UnimplementedGatewayAPIServer

	processor          *engine.Processor
	store              storage.Store
	jobNotifier        *JobNotifier
	deployNotifier     *DeployNotifier
	completionNotifier *CompletionNotifier
}

// NewServer creates a new gateway server.
func NewServer(processor *engine.Processor, store storage.Store, jobNotifier *JobNotifier, deployNotifier *DeployNotifier, completionNotifier *CompletionNotifier) *Server {
	return &Server{
		processor:          processor,
		store:              store,
		jobNotifier:        jobNotifier,
		deployNotifier:     deployNotifier,
		completionNotifier: completionNotifier,
	}
}
