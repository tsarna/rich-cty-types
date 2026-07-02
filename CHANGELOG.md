# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-07-01

### Added

- `IsContextObject(cty.Value) error` predicate wrapping `GetContextFromValue`
  (nil on success, the extraction error otherwise). Suitable for a host's
  open-type registration (e.g. functy's `Parser.RegisterOpenType("ctx", …)`),
  naming the context type in type annotations while passing values through
  untouched.

## [0.3.0] - 2026-06-25

### Added

- `Deletable` interface and `delete([ctx,] thing [, key...])` generic function
  for removing entries. No arguments means "delete everything", a single
  argument is a key, and multiple arguments are reserved for addressing a nested
  key path on types that nest (a flat type may reject a path longer than one).
- `ContextObjectBuilder.ContextPointer()` returns a stable `*context.Context`
  pointing at the builder's own context field. Context-mutating capsules can
  share this pointer so a write through it (e.g. `*p = context.WithValue(*p,
  …)`) is observed by every later `GetContextFromValue` on the same builder.

### Changed

- `ContextObjectBuilder.Build()` now routes the `_ctx` capsule through
  `ContextPointer()`, so it aliases the builder's context field rather than a
  by-value copy. Existing consumers are unaffected — they still dereference to a
  `context.Context` — but in-place context mutations are now observable through
  the built object. Not a breaking API change.

## [0.2.0] - 2026-05-29

### Added

- `Toggleable` interface and `toggle([ctx,] thing [, ...])` generic function for
  flipping a value. Returns the new value after toggling.

### Changed

- License changed to Apache-2.0
- **Breaking:** `Watcher.OnChange` now takes a `source Watchable` parameter
  identifying which Watchable fired the notification. This lets a single
  Watcher registered with multiple Watchables disambiguate without wrapping
  itself in per-source adapters.
- **Breaking:** `WatchableMixin.NotifyAll` takes a corresponding `source
  Watchable` parameter; embedding types should pass `self`.

## [0.1.0] - 2026-04-13

- Initial release as a standalone module.
