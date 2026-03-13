// Package feel provides FEEL expression evaluation for the processing engine.
package feel

import (
	"fmt"
	"strings"

	feelpkg "github.com/pbinitiative/feel"
)

// Eval evaluates a FEEL expression against a scope of variables.
// Expressions starting with "=" have the prefix stripped (Zeebe convention).
// Returns the evaluation result (normalized to Go native types) or an error.
func Eval(expression string, scope map[string]any) (any, error) {
	expression = strings.TrimPrefix(expression, "=")
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}
	result, err := feelpkg.EvalStringWithScope(expression, feelpkg.Scope(scope))
	if err != nil {
		return nil, err
	}
	return normalize(result), nil
}

// normalize converts FEEL-specific types to Go native types.
func normalize(v any) any {
	switch val := v.(type) {
	case *feelpkg.Number:
		// If the number is an integer, return int64; otherwise float64
		f := val.Float64()
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = normalize(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = normalize(item)
		}
		return out
	default:
		return v
	}
}

// EvalBool evaluates a FEEL expression and converts the result to bool.
func EvalBool(expression string, scope map[string]any) (bool, error) {
	result, err := Eval(expression, scope)
	if err != nil {
		return false, err
	}
	switch v := result.(type) {
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("expression %q did not evaluate to boolean, got %T: %v", expression, result, result)
	}
}
