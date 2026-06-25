# rich-cty-types

Generic capability interfaces, dispatcher functions, and `_ctx` / `_capsule` rich-object helpers for [go-cty](https://github.com/zclconf/go-cty).

[![CI](https://github.com/tsarna/rich-cty-types/actions/workflows/ci.yml/badge.svg)](https://github.com/tsarna/rich-cty-types/actions/workflows/ci.yml)

## Overview

When building HCL2- or cty-based DSLs, it's common to wrap Go values in cty capsules so expressions can operate on them. This package provides three reusable layers on top of go-cty:

1. **Capability interfaces** — a small set of Go interfaces (`Stringable`, `Gettable`, `Settable`, `Watchable`, …) that a wrapped value may implement to opt into generic operations.
2. **"Rich object" convention** — cty objects carrying a `_capsule` attribute (the encapsulated value) and optionally a `_ctx` attribute (a `context.Context` capsule). Helpers transparently unwrap either form.
3. **Dispatcher functions** — cty `function.Function`s (`get`, `set`, `call`, `tostring`, `length`, `watch`, …) that inspect the argument, pull out the capsule, type-assert to the relevant interface, and invoke it.

Extracted from [vinculum](https://github.com/tsarna/vinculum), where it powers the HCL expression language for variables, metrics, HTTP clients, conditions, and triggers.

## Capability interfaces

| Interface       | Method                                                                  | Used by                    |
|-----------------|-------------------------------------------------------------------------|----------------------------|
| `Stringable`    | `ToString(ctx) (string, error)`                                         | `tostring()`               |
| `Lengthable`    | `Length(ctx) (int64, error)`                                            | `length()`                 |
| `Callable`      | `Call(ctx, args) (cty.Value, error)`                                    | `call()`                   |
| `Gettable`      | `Get(ctx, args) (cty.Value, error)`                                     | `get()`                    |
| `Settable`      | `Set(ctx, args) (cty.Value, error)`                                     | `set()`                    |
| `Incrementable` | `Increment(ctx, args) (cty.Value, error)`                               | `increment()`, `decrement()` |
| `Observable`    | `Observe(ctx, args) (cty.Value, error)`                                 | `observe()`                |
| `Countable`     | `Count(ctx) (int64, error)`                                             | `count()`                  |
| `Resettable`    | `Reset(ctx) error`                                                      | `reset()`                  |
| `Stateful`      | `State(ctx) (string, error)`                                            | `state()`                  |
| `Clearable`     | `Clear(ctx) error`                                                      | `clear()`                  |
| `Deletable`     | `Delete(ctx, args) (cty.Value, error)`                                  | `delete()`                 |
| `Watchable`     | `Watch(Watcher)` / `Unwatch(Watcher)`                                   | change-notification        |
| `Watcher`       | `OnChange(ctx, old, new cty.Value)`                                     | receiver side              |

### `WatchableMixin`

Embed `WatchableMixin` to get thread-safe `Watch` / `Unwatch` plus a `NotifyAll(ctx, old, new)` helper to fan out change events to all registered watchers. Call `NotifyAll` *after* releasing your own value mutex.

## Context (`_ctx`) helpers

```go
import richcty "github.com/tsarna/rich-cty-types"

ctxVal := richcty.NewContextCapsule(ctx)              // cty.Value wrapping a context.Context
ctx2, err := richcty.GetContextFromValue(ctxVal)      // round-trip

// Or build a "ctx" variable as a cty object with extra attributes:
obj, err := richcty.NewContextObject(ctx).
    WithStringAttribute("user", "alice").
    WithInt64Attribute("request_id", 42).
    Build()
// obj is a cty.Object with attributes {user, request_id, _ctx} —
// pass it as evalCtx.Variables["ctx"].
```

`GetContextFromValue` accepts either a raw context capsule or an object with a `_ctx` attribute, so HCL expressions can pass `ctx` directly.

`Build()`'s `_ctx` capsule aliases the builder's own context field. Use `ContextPointer()` to get that same `*context.Context` for a context-mutating capsule, so a write through it is seen by every later `GetContextFromValue` on the built object:

```go
b := richcty.NewContextObject(ctx)
p := b.ContextPointer()                 // *context.Context, shared with _ctx
b.WithAttribute("baggage", newBaggageCapsule(p))
obj, _ := b.Build()
// a capsule that does `*p = context.WithValue(*p, …)` is now visible via obj._ctx
```

## Capsule helpers

```go
enc, err := richcty.GetCapsuleFromValue(val)  // unwraps capsule or object with _capsule
if s, ok := enc.(richcty.Stringable); ok { ... }

name, ok := richcty.GetStringAttr(obj, "name")
n, ok    := richcty.GetIntAttr(obj, "count")
f, ok    := richcty.GetFloat32Attr(obj, "ratio")
```

## Generic dispatcher functions

`GetGenericFunctions()` returns a ready-to-register map:

```go
import richcty "github.com/tsarna/rich-cty-types"

evalCtx.Functions = map[string]function.Function{}
for name, fn := range richcty.GetGenericFunctions() {
    evalCtx.Functions[name] = fn
}
```

Provides: `call`, `get`, `set`, `increment`, `decrement`, `observe`, `tostring`, `length`, `state`, `clear`, `delete`, `reset`, `count`.

Most accept an optional leading `ctx` argument — e.g. `get(ctx, thing, default)` or just `get(thing, default)` (falls back to `context.Background()`).

`tostring` and `length` fall through to `stdlib.MakeToFunc(cty.String)` / `stdlib.LengthFunc` for values that aren't `Stringable` / `Lengthable`, so they're safe drop-in replacements.

## License

BSD 2-Clause — see [LICENSE](LICENSE).
