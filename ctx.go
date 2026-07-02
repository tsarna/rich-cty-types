package richcty

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

// ContextCapsuleType is a cty capsule type for wrapping Context instances.
var ContextCapsuleType = cty.CapsuleWithOps("_context", reflect.TypeOf((*any)(nil)).Elem(), &cty.CapsuleOps{
	GoString: func(val interface{}) string {
		return fmt.Sprintf("_ctx(%p)", val)
	},
	TypeGoString: func(_ reflect.Type) string {
		return "_ctx"
	},
})

// NewContextCapsule creates a new cty capsule value wrapping a Context.
func NewContextCapsule(ctx context.Context) cty.Value {
	return cty.CapsuleVal(ContextCapsuleType, &ctx)
}

// Gets a context from a value, which must either be a context capsule
// or an object with a _ctx attribute that is a context capsule.
func GetContextFromValue(val cty.Value) (context.Context, error) {
	ok := true

	if val.Type().IsObjectType() {
		if val.Type().HasAttribute("_ctx") {
			val = val.GetAttr("_ctx")
		} else {
			ok = false
		}
	}

	if ok && val.Type() == ContextCapsuleType {
		encapsulated := val.EncapsulatedValue()
		ctx, ok := encapsulated.(*context.Context)
		if ok {
			return *ctx, nil
		}
	}

	return nil, fmt.Errorf("expected context capsule or object with context as ._ctx, got %s", val.Type().FriendlyName())
}

// IsContextObject reports whether val satisfies the "ctx" open type: a context
// capsule, or an object carrying a context capsule under a _ctx attribute. It
// returns nil when val is a context, else the descriptive error from
// GetContextFromValue. It is the RegisterOpenType predicate a host uses to name
// this type in functy annotations, and passes the value through untouched.
func IsContextObject(val cty.Value) error {
	_, err := GetContextFromValue(val)
	return err
}

// ContextObjectBuilder builds a cty object value for use as the "ctx" variable.
type ContextObjectBuilder struct {
	Ctx        context.Context
	Attributes map[string]cty.Value
}

// NewContextObject creates a new ContextObjectBuilder wrapping the given context.
func NewContextObject(ctx context.Context) *ContextObjectBuilder {
	return &ContextObjectBuilder{
		Ctx:        ctx,
		Attributes: make(map[string]cty.Value),
	}
}

func (b *ContextObjectBuilder) WithAttribute(name string, value cty.Value) *ContextObjectBuilder {
	b.Attributes[name] = value
	return b
}

func (b *ContextObjectBuilder) WithInt64Attribute(name string, value int64) *ContextObjectBuilder {
	b.Attributes[name] = cty.NumberIntVal(value)
	return b
}

func (b *ContextObjectBuilder) WithUInt64Attribute(name string, value uint64) *ContextObjectBuilder {
	b.Attributes[name] = cty.NumberUIntVal(value)
	return b
}

func (b *ContextObjectBuilder) WithStringAttribute(name string, value string) *ContextObjectBuilder {
	b.Attributes[name] = cty.StringVal(value)
	return b
}

// ContextPointer returns a stable pointer to the builder's context, suitable
// for context-mutating capsules that must alias the _ctx capsule's cell. Both
// the _ctx capsule (see Build) and any such capsule route through this pointer,
// so a write through it (e.g. *p = context.WithValue(*p, …)) is observed by
// every later GetContextFromValue on the same builder.
func (b *ContextObjectBuilder) ContextPointer() *context.Context {
	return &b.Ctx
}

// Build creates the cty object value for the "ctx" variable.
func (b *ContextObjectBuilder) Build() (cty.Value, error) {
	// Route _ctx through ContextPointer so it aliases &b.Ctx rather than a
	// fresh by-value copy. This lets context-mutating capsules share the same
	// cell and have their writes observed by later GetContextFromValue calls.
	b.Attributes["_ctx"] = cty.CapsuleVal(ContextCapsuleType, b.ContextPointer())
	return cty.ObjectVal(b.Attributes), nil
}
