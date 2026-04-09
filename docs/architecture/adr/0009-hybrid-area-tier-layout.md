# ADR-0009: Hybrid Area-Tier Directory Layout

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

With 16 bounded contexts, a flat directory under `internal/context/` becomes hard to
navigate. Developers need to quickly identify which team owns a BC, how critical it
is (core vs. supporting), and which BCs are related. The directory structure should
encode these relationships without requiring external documentation.

## Decision

Use a three-level hierarchy: `internal/context/<area>/<tier>/<bc>/`.

**Areas** group BCs by product domain:
- `iam/` -- identity, access, sessions, audit
- `admin/` -- platform configuration and operations
- `content/` -- user-facing content and communication
- `ops/` -- observability, rate limiting, system health

**Tiers** classify BCs by strategic importance (DDD strategic patterns):
- `core/` -- competitive advantage; highest investment (currently reserved)
- `generic/` -- common capabilities shared across products
- `supporting/` -- necessary but not differentiating

For example:
```
internal/context/iam/generic/user/
internal/context/iam/supporting/audit/
internal/context/admin/generic/featureflag/
internal/context/ops/supporting/iprule/
```

Reclassifying a BC (e.g., promoting `audit` from supporting to generic) is a single
`git mv` between tier directories within the same area.

## Consequences

### Positive
- Directory path communicates ownership (area) and investment level (tier) at a
  glance.
- IDE file trees group related BCs together, reducing cognitive load.
- Reclassification is a rename, not a restructure.

### Negative
- Four directory levels before reaching code (`internal/context/iam/generic/user/domain/`)
  creates long import paths.
- `core/` tier directories are currently empty, which may confuse newcomers.
- Area boundaries may shift as the product evolves, requiring `git mv` of entire
  subtrees.

## Alternatives Considered

- **Tier-first layout** (`internal/context/core/<bc>/`) -- groups BCs by importance
  but loses area locality; unrelated BCs sit side by side.
- **Flat layout** (`internal/context/<bc>/`) -- simple but provides no organisational
  signal; 16 directories in one folder is unwieldy.
- **Package-per-layer** (`internal/domain/`, `internal/application/`) -- standard Go
  layout but forces all BCs to share a package namespace, breaking encapsulation.
