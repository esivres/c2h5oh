package intent

// SetVariablesIntent requests update of variables in a scope.
type SetVariablesIntent struct {
	Variables []byte
	Header
	ScopeKey uint64
}

func (i *SetVariablesIntent) IntentType() Type { return SetVariables }
