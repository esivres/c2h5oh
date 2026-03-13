// Package storage defines interfaces for the BPMN engine persistence layer.
//
// Storage is organized as a set of small, focused repositories.
// Each repository handles one entity type.
//
// Unit of work follows a functional pattern inspired by net/http Handler/HandlerFunc:
//
//	handler := storage.HandlerFunc(func(ctx context.Context, s storage.Store) ([]intent.Intent, error) {
//	    def, err := s.ProcessDefinitions().FindByKey(ctx, key)
//	    // ...
//	    return intents, nil
//	})
//	results, err := store.Execute(ctx, handler)
package storage

import (
	"context"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
)

// Store provides access to all repositories within a single unit of work.
// Implementations must ensure that all operations within a single Execute call
// are atomic — either all succeed or all are rolled back.
type Store interface {
	ProcessDefinitions() ProcessDefinitionRepository
	ProcessInstances() ProcessInstanceRepository
	Variables() VariableRepository
	Jobs() JobRepository
	Timers() TimerRepository
	MessageSubscriptions() MessageSubscriptionRepository
	Incidents() IncidentRepository
	Forms() FormRepository

	// Execute runs a Handler within a transactional unit of work.
	// All repository operations inside the handler share the same transaction.
	// If the handler returns an error, the transaction is rolled back.
	// Returned intents are the follow-up work to be scheduled.
	Execute(ctx context.Context, h Handler) ([]intent.Intent, error)
}

// Handler processes a unit of work against the store.
// Returns follow-up intents and an optional error.
type Handler interface {
	Handle(ctx context.Context, s Store) ([]intent.Intent, error)
}

// HandlerFunc is an adapter to allow use of ordinary functions as Handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(ctx context.Context, s Store) ([]intent.Intent, error)

func (f HandlerFunc) Handle(ctx context.Context, s Store) ([]intent.Intent, error) {
	return f(ctx, s)
}
