# ADR-0004: Hexagonal (Ports and Adapters) Architecture

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

The GCT backend must remain testable in isolation and flexible enough to swap
infrastructure components (e.g., switch cache backends, replace notification
providers) without rewriting business logic. A direct dependency from domain code to
PostgreSQL or Redis makes unit tests slow and couples business rules to
infrastructure details.

## Decision

Each bounded context follows hexagonal architecture with four layers:

| Layer            | Directory          | Depends on          |
|------------------|--------------------|---------------------|
| Domain           | `domain/`          | nothing (pure Go)   |
| Application      | `application/`     | domain ports        |
| Infrastructure   | `infrastructure/`  | domain + application ports |
| Interfaces       | `interfaces/`      | application layer   |

- **Domain** defines aggregates, value objects, and *port interfaces* (e.g.,
  `UserRepository`, `EventPublisher`).
- **Application** implements use-case orchestration via command/query handlers,
  depending only on port interfaces.
- **Infrastructure** provides adapters: PostgreSQL repositories, Redis caches,
  SMTP senders, event bus publishers.
- **Interfaces** exposes HTTP handlers (Gin) and gRPC endpoints that call
  application handlers.

The dependency rule is strictly inward: `interfaces -> application -> domain`,
with `infrastructure` implementing domain ports.

## Consequences

### Positive
- Domain logic is testable with in-memory fakes -- no database required.
- Swapping an adapter (e.g., MinIO to S3) requires only a new adapter, no domain
  changes.
- Enforced separation prevents infrastructure leaks into business rules.

### Negative
- Every external dependency requires an interface + adapter pair, adding indirection.
- Newcomers must understand the port/adapter vocabulary to navigate the codebase.

## Alternatives Considered

- **Traditional layered architecture** (controller -> service -> repository) -- allows
  upward dependencies and infrastructure to leak into services; harder to test.
- **Clean Architecture (Uncle Bob)** -- very similar in spirit but prescribes
  additional entity/use-case/presenter layers that add ceremony without clear benefit
  over the four-layer split used here.
