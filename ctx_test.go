package richcty

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ctxKey string

// TestContextPointerAliasing verifies that the _ctx capsule produced by Build
// aliases the builder's context cell, so a write through ContextPointer is
// observed by a later GetContextFromValue on the same built object.
func TestContextPointerAliasing(t *testing.T) {
	const key ctxKey = "k"

	b := NewContextObject(context.Background())
	p := b.ContextPointer()

	obj, err := b.Build()
	require.NoError(t, err)

	// Initially absent.
	got, err := GetContextFromValue(obj)
	require.NoError(t, err)
	assert.Nil(t, got.Value(key))

	// Mutate the shared cell through the pointer, as a context-mutating capsule
	// (e.g. baggage) would.
	*p = context.WithValue(*p, key, "v")

	// The same built object now observes the mutation.
	got, err = GetContextFromValue(obj)
	require.NoError(t, err)
	assert.Equal(t, "v", got.Value(key))

	// And the pointer indeed points at the builder's own field.
	assert.Same(t, &b.Ctx, p)
}
