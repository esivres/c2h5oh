package intent

// SetVariablesIntent requests update of variables in a scope.
type SetVariablesIntent struct {
	Header

	// ScopeKey is the element instance key that defines the target scope.
	ScopeKey uint64

	// Variables is the variables to set (JSON bytes).
	Variables []byte
}

func (i *SetVariablesIntent) IntentType() Type { return SetVariables }
