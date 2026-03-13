// Package behavior defines how intents are processed against storage.
//
// Each intent type maps to a Behavior function via Registry. The partition loop:
//  1. Dequeues an intent
//  2. Looks up its behavior from the registry
//  3. Wraps it into a storage.Handler
//  4. Calls store.Execute(ctx, handler) — atomic transaction
//
// Behavior receives the intent data + store, returns follow-up intents.
package behavior

import (
	"context"
	"fmt"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// Behavior processes an intent against the store and returns follow-up intents.
type Behavior func(ctx context.Context, s storage.Store, i intent.Intent) ([]intent.Intent, error)

// Typed wraps a type-specific behavior function into a generic Behavior.
// Performs type assertion at runtime.
//
//	registry.Register(intent.DeployProcess, Typed(deployProcess))
func Typed[T intent.Intent](fn func(ctx context.Context, s storage.Store, i T) ([]intent.Intent, error)) Behavior {
	return func(ctx context.Context, s storage.Store, i intent.Intent) ([]intent.Intent, error) {
		typed, ok := i.(T)
		if !ok {
			return nil, fmt.Errorf("behavior: expected %T, got %T", *new(T), i)
		}
		return fn(ctx, s, typed)
	}
}

// AsHandler wraps a Behavior call with a specific intent into a storage.Handler.
// This bridges behavior → storage.Execute for transactional execution.
func AsHandler(b Behavior, i intent.Intent) storage.Handler {
	return storage.HandlerFunc(func(ctx context.Context, s storage.Store) ([]intent.Intent, error) {
		return b(ctx, s, i)
	})
}

// ElementTyped is an optional interface for intents that carry an element type
// for secondary dispatch in the registry.
type ElementTyped interface {
	GetElementType() string
}

// Registry maps intent types to their behavior functions.
// Supports composite keys (intentType + elementType) for secondary dispatch.
type Registry struct {
	behaviors map[intent.Type]Behavior
	typed     map[string]Behavior // key: "INTENT_TYPE:elementType"
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		behaviors: make(map[intent.Type]Behavior),
		typed:     make(map[string]Behavior),
	}
}

// Register adds a behavior for a given intent type (fallback).
func (r *Registry) Register(t intent.Type, b Behavior) {
	r.behaviors[t] = b
}

// RegisterWithElementType adds a behavior for a (intentType, elementType) pair.
func (r *Registry) RegisterWithElementType(t intent.Type, elementType string, b Behavior) {
	r.typed[string(t)+":"+elementType] = b
}

// Lookup returns the behavior for a given intent type.
// Returns nil if not registered.
func (r *Registry) Lookup(t intent.Type) Behavior {
	return r.behaviors[t]
}

// LookupWithElementType returns the behavior for (intentType, elementType).
// Falls back to the base intent type registration if no composite match.
func (r *Registry) LookupWithElementType(t intent.Type, elementType string) Behavior {
	if b := r.typed[string(t)+":"+elementType]; b != nil {
		return b
	}
	return r.behaviors[t]
}

// MustLookup returns the behavior or panics if not registered.
func (r *Registry) MustLookup(t intent.Type) Behavior {
	b := r.behaviors[t]
	if b == nil {
		panic(fmt.Sprintf("behavior: no behavior registered for intent type %q", t))
	}
	return b
}
