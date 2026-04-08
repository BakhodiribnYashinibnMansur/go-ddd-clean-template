# ADR-0005: In-Memory Event Bus with Outbox Pattern

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

Bounded contexts must react to changes in other BCs (e.g., the audit BC logs events
when a user is created in the IAM BC). Direct function calls between BCs would
create import cycles and tight coupling. An external message broker adds operational
complexity that is premature for the current deployment model (single binary).

## Decision

Use a two-tier event system:

1. **In-process event bus** (`internal/kernel/infrastructure/eventbus/inmemory.go`)
   for synchronous fan-out within the same process. Handlers subscribe by event name
   and are invoked during `Publish()`. A Redis Streams adapter
   (`redis_streams.go`) exists for future multi-instance deployments.

2. **Outbox pattern** (`internal/kernel/outbox/`) for guaranteed delivery. Commands
   write an outbox entry in the same database transaction as the aggregate mutation.
   A `Relay` goroutine polls the outbox table every 2 seconds (configurable), batch
   size 100, and forwards entries to the event bus via `RawPublisher`.

Event contracts live in `internal/contract/events/` with typed structs, an envelope
wrapper, and a name-to-struct mapping for deserialization.

## Consequences

### Positive
- Zero infrastructure beyond PostgreSQL for reliable event delivery.
- Outbox + transaction guarantees at-least-once delivery without 2PC.
- Switching to Redis Streams or another broker requires only a new adapter.

### Negative
- In-memory bus delivers events synchronously, adding latency to command handlers.
- Outbox relay introduces up to 2-second delivery delay for cross-BC events.
- At-least-once semantics require idempotent handlers.

## Alternatives Considered

- **RabbitMQ / NATS from day one** -- adds broker to the deployment stack; overkill
  when all BCs run in a single process.
- **Direct function calls between BCs** -- creates import cycles and violates the
  no-cross-BC-import rule (see ADR-0012).
- **CDC (Change Data Capture)** via Debezium -- powerful but requires Kafka and
  connector infrastructure that is disproportionate to current scale.
