# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2026-09-15

### Added

- Add the optional `store.AtomicStore` interface and an atomic `Update` operation
  to `MemoryStore`.
- Add concurrency coverage for atomic memory-store updates and multiple
  `TokenBucket` instances sharing the same store.

### Fixed

- Prevent concurrent rate-limit decisions from overwriting token debits when
  multiple `TokenBucket` instances share a `MemoryStore`.
- Serialize rate-limit decisions made through a single `TokenBucket` when its
  custom store implements only the original `Store` interface.

## [1.0.0] - 2026-01-17

### Added

- Add the token-bucket rate limiter with configurable rate, duration, and burst.
- Add detailed allow results with remaining tokens, reset time, and retry delay.
- Add the thread-safe in-memory store with optional stale-entry cleanup.
- Add reset operations, validation errors, tests, benchmarks, and runnable basic
  and HTTP examples.

[Unreleased]: https://github.com/YasserElgammal/rate-limiter/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/YasserElgammal/rate-limiter/releases/tag/v1.0.0
