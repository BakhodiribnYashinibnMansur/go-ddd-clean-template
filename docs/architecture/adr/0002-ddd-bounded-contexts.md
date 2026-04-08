# ADR-0002: Domain-Driven Design with Bounded Contexts

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

The GCT backend serves multiple product areas (IAM, content, admin, ops) with
distinct domain models. A single shared model leads to tangled dependencies, naming
collisions (e.g., `User` means different things in IAM vs. audit), and teams
stepping on each other during development.

We need clear boundaries that let each domain evolve independently while sharing a
deployment unit.

## Decision

Organise the codebase into 16 bounded contexts under
`internal/context/<area>/<tier>/<bc>/`. The four areas are:

| Area      | Bounded Contexts |
|-----------|-----------------|
| `iam`     | user, session, authz, usersetting, audit |
| `admin`   | featureflag, integration, sitesetting, statistics, dataexport, errorcode |
| `content` | file, notification, translation, announcement |
| `ops`     | metric, ratelimit, systemerror, iprule |

Each BC contains exactly four layers: `domain/`, `application/`, `infrastructure/`,
`interfaces/`. BCs must never import another BC directly; all cross-BC communication
flows through domain events (see ADR-0012).

## Consequences

### Positive
- Each BC owns its aggregate roots, value objects, and repository interfaces.
- Teams can work on separate BCs in parallel with minimal merge conflicts.
- Reclassifying a BC (e.g., promoting from supporting to generic) is a `git mv`.

### Negative
- 16 BCs create a deep directory tree that can intimidate newcomers.
- Some supporting BCs (e.g., errorcode) have thin domain layers, adding boilerplate.

## Alternatives Considered

- **Feature-folders** (`internal/user/`, `internal/file/`) -- flat structure does not
  encode area ownership or strategic tier, making it hard to reason about coupling.
- **Microservices per BC** -- operational overhead (deployment, networking, data
  consistency) is too high for the current team size and traffic.
