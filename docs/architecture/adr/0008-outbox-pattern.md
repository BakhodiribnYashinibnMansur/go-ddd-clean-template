# ADR-0008: Outbox Pattern for Reliable Event Publishing

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

When a command handler persists an aggregate change and then publishes a domain event,
two things can go wrong: the database write succeeds but the event publish fails
(lost event), or the event publishes but the database write fails (phantom event).
Both break cross-BC consistency guarantees.

## Decision

Adopt the transactional outbox pattern, implemented in `internal/kernel/outbox/`:

1. **Write phase** (`entry.go`, `postgres.go`): The command handler inserts an outbox
   entry (`event_name`, `payload`, `created_at`, `published_at = NULL`) in the *same*
   database transaction as the aggregate mutation. If the transaction rolls back, the
   event is never written.

2. **Relay phase** (`relay.go`): A background goroutine (`Relay`) polls for entries
   where `published_at IS NULL`, forwards them to a `RawPublisher` interface, and
   marks them as published. Default polling interval is 2 seconds, batch size 100.

3. **Serialization** (`serialize.go`): Events are serialized to JSON using the
   contract structs from `internal/contract/events/`.

The `RawPublisher` interface decouples the relay from the bus implementation -- today
it calls the in-memory bus; tomorrow it could push to Redis Streams or Kafka.

## Consequences

### Positive
- Atomic: event is guaranteed to exist if and only if the aggregate mutation commits.
- No distributed transactions or two-phase commit required.
- Simple to implement -- only needs PostgreSQL, no additional infrastructure.

### Negative
- Delivery is at-least-once; handlers must be idempotent.
- Polling introduces latency (up to `interval` duration) before events are delivered.
- Outbox table grows over time; requires a retention/cleanup job.

## Alternatives Considered

- **Change Data Capture (CDC)** via Debezium/Kafka Connect -- captures row changes
  from the WAL, but requires Kafka infrastructure and connector management.
- **Fire-and-forget publish** after commit -- loses events on process crash between
  commit and publish; unacceptable for audit and notification flows.
- **Listen/Notify (PostgreSQL)** -- lightweight but unreliable under connection drops;
  no built-in retry or persistence.
