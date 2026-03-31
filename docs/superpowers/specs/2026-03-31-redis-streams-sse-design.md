# Redis Streams + SSE Real-time Event Infrastructure

**Date:** 2026-03-31
**Status:** Approved

## Overview

Replace `InMemoryEventBus` with `RedisStreamsEventBus` and add SSE (Server-Sent Events) endpoints for real-time push to browser/mobile clients. Redis Pub/Sub used as instant signal layer on top of persistent Redis Streams.

## Use Cases

| # | Use Case | Delivery | Stream Key | Signal Channel |
|---|----------|----------|-----------|----------------|
| 1 | Real-time notification | SSE → Client | `stream:notifications:{user_id}` | `signal:notifications:{user_id}` |
| 3 | Audit log streaming | SSE → Client | `stream:audit` | `signal:audit` |
| 4 | Feature flag sync | Internal only | `stream:featureflags` | `signal:featureflags` |
| 5 | Cache invalidation | Internal only | `stream:cache:invalidate` | `signal:cache:invalidate` |
| 6 | Job progress | SSE → Client | `stream:jobs:{job_id}` | `signal:jobs:{job_id}` |
| 8 | System monitoring | SSE → Client | `stream:monitoring` | `signal:monitoring` |

## Architecture

```
Domain Event (notification.sent, audit_log.created, ...)
      │
      ▼
RedisStreamsEventBus
      │
      ├──► XADD stream:{channel} {data}     (persistent, MAXLEN ~1000)
      └──► PUBLISH signal:{channel} "new"    (instant notification)

SSE Handler (for use cases 1, 3, 6, 8)
      │
      ├──► SUBSCRIBE signal:{channel}        (waits for signal)
      ├──► Signal received → XREAD stream    (fetch new messages)
      └──► Push to client via SSE

Internal Listeners (for use cases 4, 5)
      │
      ├──► SUBSCRIBE signal:featureflags     (all instances)
      └──► SUBSCRIBE signal:cache:invalidate (all instances)
```

### Reconnect Flow

1. Client disconnects
2. Client reconnects with `Last-Event-ID` header (Redis stream ID)
3. SSE handler does `XREAD` from that ID forward
4. All missed messages delivered immediately
5. Resumes normal signal-based flow

## SSE Endpoints

```
GET /api/v1/stream/notifications     ← auth required (user's own)
GET /api/v1/stream/audit             ← admin role required
GET /api/v1/stream/monitoring        ← admin role required
GET /api/v1/stream/jobs/:id          ← auth required
```

## File Structure

```
internal/shared/infrastructure/
  eventbus/
    inmemory.go                       ← existing (keep for dev/test)
    redis_streams.go                  ← NEW: RedisStreamsEventBus
  sse/
    hub.go                            ← NEW: SSE connection manager
    handler.go                        ← NEW: Gin SSE handlers
    routes.go                         ← NEW: SSE route registration
  pubsub/
    subscriber.go                     ← NEW: Redis Pub/Sub subscriber
    featureflag_listener.go           ← NEW: Feature flag sync (internal)
    cache_invalidation_listener.go    ← NEW: Cache invalidation (internal)
```

## Key Interfaces

### RedisStreamsEventBus

Implements existing `application.EventBus` interface:

```go
type RedisStreamsEventBus struct {
    client *redis.Client
    group  string // consumer group name (instance identifier)
}

// Publish: XADD to stream + PUBLISH signal
func (b *RedisStreamsEventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error

// Subscribe: register handler for event name, consume via XREADGROUP
func (b *RedisStreamsEventBus) Subscribe(eventName string, handler EventHandler) error
```

### SSE Hub

```go
type Hub struct {
    clients map[string]map[chan SSEMessage]bool // channel -> set of connections
    mu      sync.RWMutex
}

type SSEMessage struct {
    ID    string // Redis stream ID (used as Last-Event-ID)
    Event string // "notification", "audit", "monitoring", "job_progress"
    Data  []byte // JSON payload
}

func (h *Hub) Register(channel string) chan SSEMessage
func (h *Hub) Unregister(channel string, ch chan SSEMessage)
func (h *Hub) Broadcast(channel string, msg SSEMessage)
```

### Internal Listeners

```go
// FeatureFlagListener — subscribes to signal:featureflags
// On signal: invalidates local feature flag cache, forces re-fetch
type FeatureFlagListener struct {
    client      *redis.Client
    flagClient  *featureflag.Client
    logger      logger.Log
}

// CacheInvalidationListener — subscribes to signal:cache:invalidate
// On signal: reads stream message for {table, id}, deletes from local cache
type CacheInvalidationListener struct {
    client *redis.Client
    cache  *cache.MemoryCache
    logger logger.Log
}
```

## Stream Data Formats

### Notification Stream
```json
{"user_id": "uuid", "title": "Yangi xabar", "message": "...", "type": "INFO"}
```

### Audit Stream
```json
{"action": "USER_DELETE", "user_id": "uuid", "resource_type": "User", "resource_id": "uuid", "ip": "1.2.3.4", "timestamp": "..."}
```

### Feature Flag Stream
```json
{"key": "dark_mode", "enabled": false, "changed_by": "uuid"}
```

### Cache Invalidation Stream
```json
{"table": "users", "id": "uuid", "action": "update"}
```

### Job Progress Stream
```json
{"job_id": "uuid", "progress": 75, "status": "PROCESSING", "message": "Exporting row 7500/10000"}
```

### Monitoring Stream
```json
{"type": "system_error", "severity": "CRITICAL", "code": "DB_CONN_LOST", "message": "...", "service": "api"}
```

## Configuration

```yaml
# config.yaml
redis:
  enabled: true     # must be enabled
  host: localhost
  port: 6379

sse:
  enabled: true
  stream_max_len: 1000          # MAXLEN per stream
  heartbeat_interval: 30        # seconds, keep-alive ping
  client_buffer_size: 256       # SSE message channel buffer
```

## Stream Retention

- All streams use `MAXLEN ~1000` (approximate trimming)
- Old messages auto-evicted when limit reached
- Sufficient for reconnect replay (last ~1000 events per channel)

## Security

- SSE endpoints protected by existing `authMW` and `authzMW`
- Notification stream filtered by authenticated user's ID
- Audit and monitoring streams require admin role
- Job progress requires auth (any authenticated user for their jobs)

## Testing

- `InMemoryEventBus` kept for unit tests
- `RedisStreamsEventBus` tested with `miniredis` (already in go.mod)
- SSE handlers tested with `httptest.NewRecorder` + streaming assertions
- Internal listeners tested with `miniredis` pub/sub simulation
