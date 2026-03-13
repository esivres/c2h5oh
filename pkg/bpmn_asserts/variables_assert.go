package bpmn_asserts

import (
	"encoding/json"
	"testing"
)

// VariablesMapAssert provides fluent assertions on a map of process variables.
// Variable values are stored as JSON strings (as Zeebe exports them).
type VariablesMapAssert struct {
	t    testing.TB
	vars map[string]string
}

// ForVariables creates a new VariablesMapAssert from a variables map.
// Typically used with ProcessInstanceAssert.ExtractVariables().
func ForVariables(t testing.TB, vars map[string]string) *VariablesMapAssert {
	t.Helper()
	return &VariablesMapAssert{t: t, vars: vars}
}

// ContainsVariable asserts that a variable with the given name exists.
func (a *VariablesMapAssert) ContainsVariable(name string) *VariablesMapAssert {
	a.t.Helper()
	if _, ok := a.vars[name]; !ok {
		a.t.Fatalf("expected variables to contain %q, but available variables are %v",
			name, variableNames(a.vars))
	}
	return a
}

// DoesNotContainVariable asserts that a variable with the given name does NOT exist.
func (a *VariablesMapAssert) DoesNotContainVariable(name string) *VariablesMapAssert {
	a.t.Helper()
	if _, ok := a.vars[name]; ok {
		a.t.Fatalf("expected variables NOT to contain %q, but it exists with value %s",
			name, a.vars[name])
	}
	return a
}

// HasVariableWithValue asserts that a variable exists with the given name and value.
// The expected value is JSON-compared.
func (a *VariablesMapAssert) HasVariableWithValue(name string, expectedValue any) *VariablesMapAssert {
	a.t.Helper()
	actualJSON, ok := a.vars[name]
	if !ok {
		a.t.Fatalf("expected variables to contain %q, but available variables are %v",
			name, variableNames(a.vars))
		return a
	}

	expectedJSON, err := json.Marshal(expectedValue)
	if err != nil {
		a.t.Fatalf("failed to marshal expected value for variable %q: %v", name, err)
		return a
	}

	if !jsonEqual(actualJSON, string(expectedJSON)) {
		a.t.Fatalf("expected variable %q to have value %s, but was %s",
			name, string(expectedJSON), actualJSON)
	}
	return a
}

// HasSize asserts that the variables map contains exactly the specified number of entries.
func (a *VariablesMapAssert) HasSize(expectedSize int) *VariablesMapAssert {
	a.t.Helper()
	if len(a.vars) != expectedSize {
		a.t.Fatalf("expected variables to have %d entries, but has %d: %v",
			expectedSize, len(a.vars), variableNames(a.vars))
	}
	return a
}

// IsEmpty asserts that the variables map is empty.
func (a *VariablesMapAssert) IsEmpty() *VariablesMapAssert {
	a.t.Helper()
	if len(a.vars) != 0 {
		a.t.Fatalf("expected variables to be empty, but has %d entries: %v",
			len(a.vars), variableNames(a.vars))
	}
	return a
}

// Raw returns the underlying variables map for custom assertions.
func (a *VariablesMapAssert) Raw() map[string]string {
	return a.vars
}
