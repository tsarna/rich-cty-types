package richcty

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noExtern are the functions deliberately left out of externs.cty: their cty
// metadata says everything true about them (one parameter, no variadic, a fixed
// return type), so an extern would only be a second place for it to drift.
var noExtern = map[string]bool{"length": true, "tostring": true}

// externDeclRE matches a top-level declaration in externs.cty. The file is parsed
// here with a regex rather than with functy on purpose: this package must not depend
// on functy (its bytes are opaque to it), and the check only needs the name set.
var externDeclRE = regexp.MustCompile(`(?m)^func (\w+)\(`)

// TestExternsCoverEveryFunction is the drift guard. Adding a function to
// GetGenericFunctions without declaring it in externs.cty would leave it reflecting
// as the cty lie `f(thing, ...args)`; declaring one that no longer exists would
// document a function nobody can call. Either way, this fails.
func TestExternsCoverEveryFunction(t *testing.T) {
	declared := make(map[string]bool)
	for _, m := range externDeclRE.FindAllStringSubmatch(string(Externs()), -1) {
		declared[m[1]] = true
	}

	for name := range GetGenericFunctions() {
		if noExtern[name] {
			assert.False(t, declared[name],
				"%s() is listed in noExtern but externs.cty declares it: pick one", name)
			continue
		}
		assert.True(t, declared[name],
			"%s() is provided by GetGenericFunctions but has no declaration in externs.cty, "+
				"so it reflects as the cty signature `%s(thing, ...args)` instead of its real one", name, name)
	}

	funcs := GetGenericFunctions()
	for name := range declared {
		assert.Contains(t, funcs, name,
			"externs.cty declares %s(), which GetGenericFunctions does not provide", name)
	}
}

// The bytes must declare themselves an extern file: functy's RegisterExterns
// verifies the directive rather than forcing the mode, so that this same file is a
// valid standalone .cty that `functy fmt` and `functy symbols` can open.
func TestExternsCarryTheDirective(t *testing.T) {
	require.True(t, strings.HasPrefix(string(Externs()), "//functy:extern\n"),
		"externs.cty must begin with the //functy:extern directive")
}

// Every function, and every parameter of every function, must carry a cty
// description — extern or not.
//
// The cty metadata is the only documentation a non-functy cty host can see, and the
// only thing functy's own doc() reads (doc() does not consult the extern), so a gap
// here reads as "exists but undocumented" even where help() shows a full block. An
// extern says what a signature *is*; it does not excuse the metadata from saying what
// the function does.
//
// For the twelve declared in externs.cty the cty parameters are, unavoidably, not the
// real ones — `thing` holds the *context* when one is passed, and the thing shifts
// right. Their descriptions say so, rather than pretending otherwise.
func TestEverythingIsDescribed(t *testing.T) {
	for name, fn := range GetGenericFunctions() {
		assert.NotEmpty(t, fn.Description(), "%s() has no cty Description", name)

		for _, p := range fn.Params() {
			assert.NotEmpty(t, p.Description, "%s() parameter %q has no Description", name, p.Name)
		}
		if vp := fn.VarParam(); vp != nil {
			assert.NotEmpty(t, vp.Description, "%s() variadic parameter %q has no Description", name, vp.Name)
		}
	}
}

// The two functions with no extern must document their parameters in cty, since
// nothing else does it for them.
func TestNoExternFunctionsDocumentTheirParams(t *testing.T) {
	for name := range noExtern {
		fn, ok := GetGenericFunctions()[name]
		require.True(t, ok, "%s() is not provided", name)

		params := fn.Params()
		require.Len(t, params, 1, "%s() should take exactly one parameter", name)
		assert.NotEmpty(t, params[0].Description,
			"%s() carries no extern, so its cty parameter metadata is the only documentation "+
				"there is — and %q has no Description", name, params[0].Name)
		assert.Nil(t, fn.VarParam(),
			"%s() has grown a VarParam, so its cty signature can no longer be honest; "+
				"it now needs an extern in externs.cty", name)
	}
}
