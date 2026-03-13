package behavior

import (
	"context"
	"testing"

	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"

	"github.com/stretchr/testify/require"
)

// executeCompletion runs the two-phase completion:
// phase 1 (completeElement) → ElementCompletedIntent → phase 2 (type-specific behavior).
// Returns the follow-up intents from phase 2.
func executeCompletion(t *testing.T, store storage.Store, i *intent.CompleteElementIntent) []intent.Intent {
	t.Helper()
	ctx := context.Background()
	reg := DefaultRegistry()

	// Phase 1: completeElement → returns ElementCompletedIntent
	phase1, err := store.Execute(ctx, AsHandler(
		Typed(completeElement),
		i,
	))
	require.NoError(t, err)
	require.Len(t, phase1, 1, "completeElement should return exactly one ElementCompletedIntent")

	completed, ok := phase1[0].(*intent.ElementCompletedIntent)
	require.True(t, ok, "expected ElementCompletedIntent, got %T", phase1[0])

	// Phase 2: dispatch to type-specific behavior
	b := reg.LookupWithElementType(intent.ElementCompleted, completed.ElementType)
	require.NotNil(t, b, "no behavior for ElementCompleted:%s", completed.ElementType)

	phase2, err := store.Execute(ctx, AsHandler(b, completed))
	require.NoError(t, err)
	return phase2
}

// executeActivation runs the two-phase activation:
// phase 1 (activateElement) → ElementActivatedIntent → phase 2 (type-specific behavior).
// Returns the follow-up intents from phase 2.
func executeActivation(t *testing.T, store storage.Store, i *intent.ActivateElementIntent) []intent.Intent {
	t.Helper()
	ctx := context.Background()
	reg := DefaultRegistry()

	// Phase 1: activateElement → returns ElementActivatedIntent
	phase1, err := store.Execute(ctx, AsHandler(
		Typed(activateElement),
		i,
	))
	require.NoError(t, err)
	require.Len(t, phase1, 1, "activateElement should return exactly one ElementActivatedIntent")

	activated, ok := phase1[0].(*intent.ElementActivatedIntent)
	require.True(t, ok, "expected ElementActivatedIntent, got %T", phase1[0])

	// Phase 2: dispatch to type-specific behavior
	b := reg.LookupWithElementType(intent.ElementActivated, activated.ElementType)
	require.NotNil(t, b, "no behavior for ElementActivated:%s", activated.ElementType)

	phase2, err := store.Execute(ctx, AsHandler(b, activated))
	require.NoError(t, err)
	return phase2
}
