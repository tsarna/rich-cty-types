# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
