package storage

import "context"

// Variable is a mutable entity representing a process variable within a scope.
type Variable struct {
	// Key is the unique identifier.
	Key uint64

	// ProcessInstanceKey links to the owning process instance.
	ProcessInstanceKey uint64

	// ScopeKey is the element instance key that defines this variable's scope
	// (process instance, subprocess, multi-instance body).
	ScopeKey uint64

	// Name is the variable name.
	Name string

	// Value is the serialized variable value (JSON bytes).
	Value []byte
}

// VariableRepository manages process variable storage.
type VariableRepository interface {
	// Create stores a new variable.
	Create(ctx context.Context, v *Variable) error

	// Update updates an existing variable's value.
	Update(ctx context.Context, key uint64, value []byte) error

	// FindByScope returns all variables for a given scope.
	FindByScope(ctx context.Context, scopeKey uint64) ([]*Variable, error)

	// FindByName returns a variable by name within a scope.
	// Returns nil, nil if not found.
	FindByName(ctx context.Context, scopeKey uint64, name string) (*Variable, error)

	// CopyToScope copies variables from sourceScope to targetScope.
	// Used when entering a new scope (subprocess, multi-instance).
	CopyToScope(ctx context.Context, sourceScopeKey, targetScopeKey uint64, names []string) error

	// PropagateToParent copies variables from childScope to parentScope.
	// Used when exiting a scope (output mappings).
	PropagateToParent(ctx context.Context, childScopeKey, parentScopeKey uint64, names []string) error
}
