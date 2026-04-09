# ADR-0012: Cross-BC Communication via Contracts

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

Bounded contexts must remain independently deployable (in principle) and free of
import cycles. Yet BCs need to collaborate: the audit BC must know when a user is
created, the notification BC must react to session events, and the statistics BC
must aggregate data from multiple domains.

Direct Go imports between BCs would create compile-time coupling, merge conflicts,
and make it impossible to extract a BC into a separate service later.

## Decision

All cross-BC communication flows through two contract packages:

### `internal/contract/events/`
- `base_event.go` -- base struct with ID, timestamp, correlation ID.
- `envelope.go` -- wrapper carrying event name + serialized payload.
- `user_events.go` -- typed event structs (e.g., `UserCreatedEvent`,
  `UserDeactivatedEvent`).
- `user_mapping.go` -- maps event names to struct types for deserialization.

BCs publish events via the kernel's `EventBus` interface. Subscribing BCs register
handlers by event name -- they depend on the contract package, never on the
publishing BC.

### `internal/contract/ports/`
- `user_auth.go` -- port interfaces for cross-BC queries (e.g., resolving a user
  ID from a token) that cannot be modelled as events.

Port interfaces are implemented by the owning BC's infrastructure layer and injected
at startup. The consuming BC depends only on the interface, not the implementation.

### Rules
1. BCs must **never** import another BC's packages.
2. Shared types live in `internal/contract/`, not in any BC.
3. New cross-BC interactions require adding to the contract package and updating
   this ADR.

## Consequences

### Positive
- Zero compile-time coupling between BCs.
- Adding a new subscriber is additive -- no changes to the publisher.
- Contract package serves as living documentation of all BC interactions.

### Negative
- Indirection makes it harder to trace a request across BCs via IDE navigation.
- Contract package must be carefully governed to avoid becoming a dumping ground.
- Event-driven communication introduces eventual consistency that callers must
  handle.

## Alternatives Considered

- **Shared database tables** -- BCs query each other's tables directly; creates
  hidden coupling, schema lock-in, and makes extraction impossible.
- **REST/gRPC between BCs** -- adds network hops and serialization overhead within
  a single process; appropriate only after extracting to separate services.
- **Direct Go imports with interface boundaries** -- technically possible but
  Go's import graph creates transitive coupling; a change in BC-A's types forces
  recompilation of BC-B.
