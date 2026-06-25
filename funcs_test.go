package richcty

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

// deletableThing is a test capsule implementing Deletable: it records the keys
// passed to Delete.
type deletableThing struct{ deleted []string }

var deletableCapsuleType = cty.CapsuleWithOps("deletableThing", reflect.TypeOf(deletableThing{}), &cty.CapsuleOps{})

func (d *deletableThing) Delete(_ context.Context, args []cty.Value) (cty.Value, error) {
	for _, a := range args {
		d.deleted = append(d.deleted, a.AsString())
	}
	return cty.NullVal(cty.DynamicPseudoType), nil
}

func TestDeleteFunction_DispatchesToDeletable(t *testing.T) {
	fn := GetGenericFunctions()["delete"]
	require.NotNil(t, fn)

	thing := &deletableThing{}
	cap := cty.CapsuleVal(deletableCapsuleType, thing)

	got, err := fn.Call([]cty.Value{cap, cty.StringVal("a"), cty.StringVal("b")})
	require.NoError(t, err)
	assert.True(t, got.IsNull())
	assert.Equal(t, []string{"a", "b"}, thing.deleted)
}

func TestDeleteFunction_UnsupportedType(t *testing.T) {
	fn := GetGenericFunctions()["delete"]
	_, err := fn.Call([]cty.Value{newTestCapsule("x"), cty.StringVal("a")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support delete()")
}
