# ADR-0003: CQRS Pattern in Application Layer

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

Many bounded contexts expose operations that are fundamentally different in nature:
commands that mutate state (create user, revoke session) and queries that read state
(list users, get settings). Mixing both in a single service struct leads to methods
with conflicting concerns -- commands need validation, events, and transactions while
queries need pagination, filtering, and caching.

## Decision

Each BC's `application/` layer separates command handlers from query handlers:

```
application/
  command/
    create_user.go
    update_user.go
  query/
    get_user.go
    list_users.go
```

- **Command handlers** accept a command DTO, validate it, call domain logic, persist
  via repository, and publish domain events.
- **Query handlers** accept a query DTO, read from the repository (potentially a
  read-optimised view), and return a response DTO.

Commands and queries share the same PostgreSQL database and repository interfaces
defined in `domain/`. There is no separate read model or projection store.

## Consequences

### Positive
- Single-responsibility per handler makes unit testing straightforward.
- Adding a new operation is additive -- no risk of breaking existing handlers.
- Query handlers can be independently optimised (e.g., adding caching) without
  touching command logic.

### Negative
- More files per BC compared to a single service struct.
- Without a separate read model, complex queries still hit the write database.

## Alternatives Considered

- **Single repository service** -- fewer files but methods grow into a god-object;
  harder to test and reason about side-effects.
- **Full event-sourced CQRS** with separate read projections -- powerful but adds
  significant complexity (event store, projectors, eventual consistency) that is not
  justified by current query patterns.
