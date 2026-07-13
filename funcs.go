package richcty

import (
	"context"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// contextAndThing extracts an optional leading context and the "thing" value from
// args. If args[0] is a context capsule/object it is used as the context and
// args[1] is the thing; otherwise context.Background() is used and args[0] is
// the thing. Returns (ctx, thing, remaining args after thing).
func contextAndThing(args []cty.Value) (context.Context, cty.Value, []cty.Value) {
	ctx, err := GetContextFromValue(args[0])
	if err == nil {
		if len(args) < 2 {
			return ctx, cty.NilVal, nil
		}
		return ctx, args[1], args[2:]
	}
	return context.Background(), args[0], args[1:]
}

func GetGenericFunctions() map[string]function.Function {
	return map[string]function.Function{
		"call":      makeCallFunction(),
		"clear":     makeClearFunction(),
		"count":     makeCountFunction(),
		"decrement": makeDecrementFunction(),
		"delete":    makeDeleteFunction(),
		"get":       makeGetFunction(),
		"increment": makeIncrementFunction(),
		"length":    makeLengthFunction(),
		"observe":   makeObserveFunction(),
		"reset":     makeResetFunction(),
		"set":       makeSetFunction(),
		"state":     makeStateFunction(),
		"toggle":    makeToggleFunction(),
		"tostring":  makeToStringFunction(),
	}
}

func extractCallable(val cty.Value) (Callable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("argument is not callable: %w", err)
	}
	c, ok := enc.(Callable)
	if !ok {
		return nil, fmt.Errorf("%s does not support call()", val.Type().FriendlyName())
	}
	return c, nil
}

// makeCallFunction returns a cty function that invokes call() on any Callable thing.
// Signature: call(ctx, thing, args...) -> response
func makeCallFunction() function.Function {
	return function.New(&function.Spec{
		// Unlike the rest of this family the context is required, so cty can express
		// the head of the signature honestly; the variadic tail is genuinely variadic.
		// externs.cty declares it anyway, to document the arguments.
		Description: "Invoke a request/response capability: an LLM client, an HTTP endpoint, an MCP tool.",
		Params: []function.Parameter{
			{
				Name: "ctx",
				Type: cty.DynamicPseudoType,
			},
			{
				Name: "thing",
				Type: cty.DynamicPseudoType,
			},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, err := GetContextFromValue(args[0])
			if err != nil {
				return cty.NilVal, fmt.Errorf("call: context error: %w", err)
			}

			callable, err := extractCallable(args[1])
			if err != nil {
				return cty.NilVal, fmt.Errorf("call: %w", err)
			}

			return callable.Call(ctx, args[2:])
		},
	})
}

// makeClearFunction returns a cty function for clear([ctx,] thing).
func makeClearFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Cancel a thing's pending state: release a latch, drop a pending transition, discard a retentive timer's accumulated time.",
		Params:      []function.Parameter{{Name: "thing", Type: cty.DynamicPseudoType}},
		VarParam:    &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:        function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, _ := contextAndThing(args)
			enc, err := GetCapsuleFromValue(thing)
			if err != nil {
				return cty.NilVal, fmt.Errorf("clear: %w", err)
			}
			c, ok := enc.(Clearable)
			if !ok {
				return cty.NilVal, fmt.Errorf("clear: %s does not support clear()", thing.Type().FriendlyName())
			}
			if err := c.Clear(ctx); err != nil {
				return cty.NilVal, fmt.Errorf("clear: %w", err)
			}
			return cty.NullVal(cty.DynamicPseudoType), nil
		},
	})
}

// makeCountFunction returns a cty function for count([ctx,] thing) → number.
func makeCountFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "How many times something has happened — a counter's running total.",
		Params:      []function.Parameter{{Name: "thing", Type: cty.DynamicPseudoType}},
		VarParam:    &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:        function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, _ := contextAndThing(args)
			enc, err := GetCapsuleFromValue(thing)
			if err != nil {
				return cty.NilVal, fmt.Errorf("count: %w", err)
			}
			c, ok := enc.(Countable)
			if !ok {
				return cty.NilVal, fmt.Errorf("count: %s does not support count()", thing.Type().FriendlyName())
			}
			n, err := c.Count(ctx)
			if err != nil {
				return cty.NilVal, fmt.Errorf("count: %w", err)
			}
			return cty.NumberIntVal(n), nil
		},
	})
}

// makeDecrementFunction returns a cty function for decrement([ctx,] thing [, delta]).
// Delta defaults to 1. Implemented as increment(thing, -delta).
func makeDecrementFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Subtract from a numeric thing.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			i, err := extractIncrementable(thing)
			if err != nil {
				return cty.NilVal, fmt.Errorf("decrement: %w", err)
			}
			delta := cty.NumberIntVal(1)
			if len(rest) > 0 {
				delta = rest[0]
				rest = rest[1:]
			}
			neg, err := stdlib.NegateFunc.Call([]cty.Value{delta})
			if err != nil {
				return cty.NilVal, fmt.Errorf("decrement: %w", err)
			}
			return i.Increment(ctx, append([]cty.Value{neg}, rest...))
		},
	})
}

func extractDeletable(val cty.Value) (Deletable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("delete: %w", err)
	}
	d, ok := enc.(Deletable)
	if !ok {
		return nil, fmt.Errorf("delete: %s does not support delete()", val.Type().FriendlyName())
	}
	return d, nil
}

// makeDeleteFunction returns a cty function for delete([ctx,] thing, keys...).
// It removes the named entries from the thing; the thing decides what "key"
// means and what to return.
func makeDeleteFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Remove entries from a thing.",
		Params:      []function.Parameter{{Name: "thing", Type: cty.DynamicPseudoType}},
		VarParam:    &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:        function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			d, err := extractDeletable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return d.Delete(ctx, rest)
		},
	})
}

func extractGettable(val cty.Value) (Gettable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	g, ok := enc.(Gettable)
	if !ok {
		return nil, fmt.Errorf("get: %s does not support get()", val.Type().FriendlyName())
	}
	return g, nil
}

// makeGetFunction returns a cty function for get([ctx,] thing [, default, ...]).
func makeGetFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Read a thing's current value.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{
			Name: "args",
			Type: cty.DynamicPseudoType,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			g, err := extractGettable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return g.Get(ctx, rest)
		},
	})
}

func extractIncrementable(val cty.Value) (Incrementable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("increment: %w", err)
	}
	i, ok := enc.(Incrementable)
	if !ok {
		return nil, fmt.Errorf("increment: %s does not support increment()", val.Type().FriendlyName())
	}
	return i, nil
}

// makeIncrementFunction returns a cty function for increment([ctx,] thing, delta [, ...]).
func makeIncrementFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Add to a numeric thing.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			i, err := extractIncrementable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return i.Increment(ctx, rest)
		},
	})
}

// makeLengthFunction returns an enhanced length() that supports Lengthable
// capsules (and objects with _capsule), falling back to stdlib length.
//
// This function carries NO extern (see externs.cty): it has one parameter, no
// variadic, and a fixed return type, so its cty metadata below already states
// everything true about it. Keep it that way — an extern here would only be a
// second place for the same facts to drift.
func makeLengthFunction() function.Function {
	fallback := stdlib.LengthFunc
	return function.New(&function.Spec{
		Description: "How many elements a value holds. Distinct from count(), which is how many times something has happened.",
		Params: []function.Parameter{{
			Name:        "v",
			Type:        cty.DynamicPseudoType,
			Description: "The value to measure: a Lengthable capsule (or a rich object carrying one), or any collection or string.",
		}},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			enc, err := GetCapsuleFromValue(args[0])
			if err == nil {
				if l, ok := enc.(Lengthable); ok {
					n, err := l.Length(context.Background())
					if err != nil {
						return cty.NilVal, fmt.Errorf("length: %w", err)
					}
					return cty.NumberIntVal(n), nil
				}
			}
			return fallback.Call(args)
		},
	})
}

func extractObservable(val cty.Value) (Observable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("observe: %w", err)
	}
	o, ok := enc.(Observable)
	if !ok {
		return nil, fmt.Errorf("observe: %s does not support observe()", val.Type().FriendlyName())
	}
	return o, nil
}

// makeObserveFunction returns a cty function for observe([ctx,] thing, value [, ...]).
func makeObserveFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Record an observation — a sample into a histogram.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			if len(rest) == 0 {
				return cty.NilVal, fmt.Errorf("observe: missing value argument")
			}
			o, err := extractObservable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return o.Observe(ctx, rest)
		},
	})
}

// makeResetFunction returns a cty function for reset([ctx,] thing).
func makeResetFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Return a thing to its initial state: zero a counter, re-arm a watchdog.",
		Params:      []function.Parameter{{Name: "thing", Type: cty.DynamicPseudoType}},
		VarParam:    &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:        function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, _ := contextAndThing(args)
			enc, err := GetCapsuleFromValue(thing)
			if err != nil {
				return cty.NilVal, fmt.Errorf("reset: %w", err)
			}
			r, ok := enc.(Resettable)
			if !ok {
				return cty.NilVal, fmt.Errorf("reset: %s does not support reset()", thing.Type().FriendlyName())
			}
			if err := r.Reset(ctx); err != nil {
				return cty.NilVal, fmt.Errorf("reset: %w", err)
			}
			return cty.NullVal(cty.DynamicPseudoType), nil
		},
	})
}

func extractSettable(val cty.Value) (Settable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("set: %w", err)
	}
	s, ok := enc.(Settable)
	if !ok {
		return nil, fmt.Errorf("set: %s does not support set()", val.Type().FriendlyName())
	}
	return s, nil
}

// makeSetFunction returns a cty function for set([ctx,] thing [, value, ...]).
func makeSetFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Update a thing's value.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			s, err := extractSettable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return s.Set(ctx, rest)
		},
	})
}

// makeStateFunction returns a cty function for state([ctx,] thing) → string.
func makeStateFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "A thing's named internal state.",
		Params:      []function.Parameter{{Name: "thing", Type: cty.DynamicPseudoType}},
		VarParam:    &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:        function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, _ := contextAndThing(args)
			enc, err := GetCapsuleFromValue(thing)
			if err != nil {
				return cty.NilVal, fmt.Errorf("state: %w", err)
			}
			s, ok := enc.(Stateful)
			if !ok {
				return cty.NilVal, fmt.Errorf("state: %s does not support state()", thing.Type().FriendlyName())
			}
			str, err := s.State(ctx)
			if err != nil {
				return cty.NilVal, fmt.Errorf("state: %w", err)
			}
			return cty.StringVal(str), nil
		},
	})
}

func extractToggleable(val cty.Value) (Toggleable, error) {
	enc, err := GetCapsuleFromValue(val)
	if err != nil {
		return nil, fmt.Errorf("toggle: %w", err)
	}
	t, ok := enc.(Toggleable)
	if !ok {
		return nil, fmt.Errorf("toggle: %s does not support toggle()", val.Type().FriendlyName())
	}
	return t, nil
}

// makeToggleFunction returns a cty function for toggle([ctx,] thing [, ...]).
func makeToggleFunction() function.Function {
	return function.New(&function.Spec{
		// The parameters below are a lie cty forces on us: an optional leading
		// context and every named trailing argument collapse into one anonymous
		// VarParam. externs.cty carries the real signature.
		Description: "Flip a boolean-valued thing, returning its new value.",
		Params: []function.Parameter{
			{Name: "thing", Type: cty.DynamicPseudoType},
		},
		VarParam: &function.Parameter{Name: "args", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx, thing, rest := contextAndThing(args)
			t, err := extractToggleable(thing)
			if err != nil {
				return cty.NilVal, err
			}
			return t.Toggle(ctx, rest)
		},
	})
}

// makeToStringFunction returns an enhanced tostring() that supports Stringable
// capsules (and objects with _capsule), falling back to stdlib conversion.
//
// Like length(), this function carries NO extern (see externs.cty): its cty metadata
// below is complete on its own.
func makeToStringFunction() function.Function {
	fallback := stdlib.MakeToFunc(cty.String)
	return function.New(&function.Spec{
		Description: "Convert a value to its string representation.",
		Params: []function.Parameter{{
			Name:        "v",
			Type:        cty.DynamicPseudoType,
			Description: "The value to convert: a Stringable capsule (or a rich object carrying one), or any value cty can convert to string.",
		}},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			enc, err := GetCapsuleFromValue(args[0])
			if err == nil {
				if s, ok := enc.(Stringable); ok {
					str, err := s.ToString(context.Background())
					if err != nil {
						return cty.NilVal, fmt.Errorf("tostring: %w", err)
					}
					return cty.StringVal(str), nil
				}
			}
			return fallback.Call(args)
		},
	})
}
