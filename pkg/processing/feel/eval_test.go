package feel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEval_SimpleExpression(t *testing.T) {
	result, err := Eval("= 1 + 2", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result)
}

func TestEval_WithScope(t *testing.T) {
	scope := map[string]any{
		"amount": 1500,
	}
	result, err := Eval("= amount > 1000", scope)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEval_StripEqualsPrefix(t *testing.T) {
	result, err := Eval("=42", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestEval_NoPrefix(t *testing.T) {
	result, err := Eval("42", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestEval_EmptyExpression(t *testing.T) {
	_, err := Eval("", nil)
	require.Error(t, err)
}

func TestEvalBool_True(t *testing.T) {
	scope := map[string]any{"x": 10}
	result, err := EvalBool("= x > 5", scope)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestEvalBool_False(t *testing.T) {
	scope := map[string]any{"x": 3}
	result, err := EvalBool("= x > 5", scope)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestEvalBool_NotBoolean(t *testing.T) {
	_, err := EvalBool("= 42", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not evaluate to boolean")
}

func TestEval_NestedScope(t *testing.T) {
	scope := map[string]any{
		"order": map[string]any{
			"id":     "ORD-123",
			"amount": 500,
		},
	}
	result, err := Eval("= order.amount", scope)
	require.NoError(t, err)
	assert.Equal(t, int64(500), result)
}
