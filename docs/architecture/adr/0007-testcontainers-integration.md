# ADR-0007: Testcontainers for Integration Testing

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

Integration tests need real infrastructure (PostgreSQL, Redis, MinIO) to catch issues
that in-memory fakes miss: SQL syntax errors, migration drift, Redis data-type
mismatches. Running tests against shared dev databases causes flaky results and
cross-contamination between test suites.

## Decision

Use `testcontainers-go` to spin up ephemeral Docker containers per test suite. The
helpers live in `test/testcontainers/`:

| File       | Container             | Purpose                        |
|------------|-----------------------|--------------------------------|
| `psql.go`  | PostgreSQL            | Schema migrations + repository tests |
| `redis.go` | Redis                 | Cache and rate-limit adapter tests |
| `minio.go` | MinIO                 | File upload/download adapter tests |

Each helper starts a container, waits for readiness, returns a connection string, and
tears down after the test. Constants in `consts.go` pin image versions so builds are
reproducible.

The `test/` directory also holds `e2e/`, `integration/`, `performance/`, and
`schemathesis/` suites that build on these containers.

## Consequences

### Positive
- Tests run against the exact same database engine as production (PostgreSQL, not
  SQLite).
- Each suite gets a pristine database -- no cross-test pollution.
- Works identically in CI and on developer machines (only Docker required).

### Negative
- First test run pulls container images, which is slow on cold caches.
- Tests require Docker daemon access; not available in all CI sandboxes.
- Container startup adds 2-5 seconds per suite compared to in-memory fakes.

## Alternatives Considered

- **docker-compose** -- requires a separate lifecycle (`docker-compose up` before
  tests, `down` after); harder to parallelise and clean up per suite.
- **SQLite for tests** -- fast but differs from PostgreSQL in syntax, type handling,
  and constraint behaviour; hides real bugs.
- **Shared dev database** -- flaky tests due to concurrent writes; hard to run
  locally without VPN.
