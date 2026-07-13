package richcty

import _ "embed"

//go:embed externs.cty
var externsCty []byte

// ExternsFilename is the name reported for the embedded declarations in
// diagnostics.
const ExternsFilename = "rich-cty-types/externs.cty"

// Externs returns the functy `//functy:extern` declarations for the functions
// GetGenericFunctions provides: their real signatures, which their cty metadata
// cannot express.
//
// Each of these functions takes an optional *leading* context, sniffed out of the
// first argument at call time. A cty parameter list can only make its trailing
// parameters optional, so the context — and every named trailing argument with it —
// is swallowed into one anonymous variadic, and the whole family reflects as the
// useless `f(thing, ...args)`. The declarations here say what they actually are, so
// that help(), generated documentation, and editor tooling can show it.
//
// The bytes are opaque to this package: it does not import functy, and nothing here
// parses them. A functy host registers them:
//
//	parser.RegisterExterns(richcty.Externs(), richcty.ExternsFilename)
//
// length() and tostring() are deliberately absent — their cty metadata is complete,
// so an extern would only be a second place for it to drift.
func Externs() []byte { return externsCty }
