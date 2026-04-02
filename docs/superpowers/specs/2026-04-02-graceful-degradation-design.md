# Graceful Degradation — External Service Failure Handling

**Date:** 2026-04-02
**Status:** Approved

## Problem

External service failures (Firebase, Telegram, Webhook, Redis) cause silent data loss, blocked requests, or cascading failures. No circuit breaker protection exists despite `circuitbreaker.Breaker` being available in the codebase.

## Approach

Individual circuit breakers per external service using the existing `circuitbreaker.Breaker`. Each component gets its own breaker, fallback strategy, and Redis-backed buffer for undelivered payloads.

## 1. Firebase FCM

### Circuit Breakers
- `firebase-mobile` — wraps `MobileClient.Send` / `SendEachForMulticast`
- `firebase-web` — wraps `WebClient.Send` / `SendEachForMulticast`
- Config: threshold=3, timeout=60s

### Fallback (circuit open)
- Full notification payload saved to Redis LIST `fcm:fallback:buffer`
- Entry contains: `token`, `title`, `body`, `data`, `fcm_type`, `timestamp`
- Background flusher runs every **2 minutes**:
  - Reads all buffered entries from Redis
  - Groups by `title` + `body` (distinct)
  - Sends **one consolidated Telegram alert** with full data per group:
    ```
    🔕 FCM unavailable (last 2min):

    1. "Yangi buyurtma" × 42 recipients
       tokens: abc12..., def34..., ghi56... (+39 more)
       data: {"order_id":"123","status":"new"}

    2. "To'lov tasdiqlandi" × 7 recipients
       tokens: xyz78..., uvw90... (+5 more)
       data: {"payment_id":"456","amount":"50000"}

    Total: 49 notifications lost
    ```
- Also logged at `Warn` level with `operation=fcm_fallback` (persists to DB via log flusher)

### SendNotifications loop fix
- Currently: one token fails → entire loop stops
- Change: skip failed token, continue loop, log total fail count at end

### Redis also down
- In-memory counter fallback, sends simple `"FCM down, [N] notifications lost"` to Telegram when it recovers

### Telegram also down (during FCM fallback flush)
- If Telegram circuit is also open when FCM flusher tries to send consolidated alert → alert goes to `telegram:pending` buffer automatically (Telegram's own fallback handles it)
- Worst case (Redis + Telegram + Firebase all down): logged at `Warn` level to DB via log persist

## 2. Telegram

### Circuit Breaker
- `telegram` breaker wraps `Client.SendMessage`
- Config: threshold=3, timeout=120s

### Fallback (circuit open)
- Message saved to Redis LIST `telegram:pending`
- Entry: `{"message_type":"error","text":"...","failed_at":"..."}`
- Background goroutine every **30s**:
  - Checks circuit state
  - If closed (Telegram recovered) → flush pending messages from Redis, send sequentially
  - If still open → skip
  - Flush failure → messages stay in Redis for next interval

### Redis also down
- Logger `Warn` with full message text (persists to DB via log flusher)

### Recovery logging
- Circuit transition open→closed: `Info` log `"telegram recovered"`

## 3. Webhook Reporter

### Circuit Breaker
- `webhook` breaker wraps HTTP POST
- Config: threshold=3, timeout=60s

### Async conversion
- Current: `SendError` makes synchronous HTTP call → blocks caller
- Change: payload goes to Redis LIST `webhook:pending` immediately (never blocks)
- Background worker sends from Redis, through circuit breaker

### Fallback (circuit open)
- Payloads stay in Redis `webhook:pending`
- Background flusher every **30s** checks circuit, sends when webhook recovers

### Redis also down
- Logger `Warn` with payload (persists to DB)

## 4. SSE Bridge

### Current problem
- `XRead` error → silent `continue`, no logging
- Pub/Sub channel closes → goroutine exits silently

### Reconnect with exponential backoff
- Subscribe fails or channel closes → reconnect loop:
  - Backoff: 1s → 2s → 4s → 8s → 16s → **max 30s**
  - Each attempt: `Warn` log `"SSE bridge reconnecting to [signalChannel]"`
  - Success: `Info` log `"SSE bridge reconnected"`, backoff reset to 1s
- `XRead` errors: log at `Warn` level (currently silent)
- `ctx.Done()` → clean exit, no reconnect

## 5. RedisStreamsEventBus

### Current problem
- `Publish` → `XAdd` fails → error return → local handlers never called → event lost

### In-memory fallback
- `XAdd` or `Publish` (pub/sub signal) fails:
  - Log `Warn` `"eventbus: redis unavailable, falling back to in-memory"` (once, on transition)
  - Skip Redis operations (XADD + PUBLISH signal)
  - **Local handlers still called** — app-internal event flow continues
- Redis recovers → next `Publish` succeeds → `Info` log `"eventbus: redis recovered"` (once, on transition)
- State tracking: `redisDown atomic.Bool` (same pattern as log persist)

### What is lost during Redis downtime
- SSE Bridge won't receive signals (real-time push stops)
- Stream persistence stops (no XADD)
- Local handlers (DB writes, cache invalidation) continue normally

## 6. Alerter

### Current problem
- `_, _ = a.enqueuer.EnqueueTask(...)` — enqueue failure silently ignored, alert lost

### Fallback
- Enqueue fails → save to Redis LIST `alerter:pending`
- Background goroutine every **30s** → re-enqueue pending alerts
- Redis also down → Logger `Warn` with full alert text (persists to DB)

### Relationship with Telegram pending
Two independent buffers:
1. **Alerter** `alerter:pending` — catches Asynq enqueue failures
2. **Telegram** `telegram:pending` — catches Telegram API failures

They don't overlap: Alerter buffer feeds into Asynq, Telegram buffer feeds into Telegram API.

## Circuit Breaker Configurations Summary

| Service | Breaker Name | Threshold | Timeout | Buffer Key | Flush Interval |
|---------|-------------|-----------|---------|------------|----------------|
| Firebase Mobile | `firebase-mobile` | 3 | 60s | `fcm:fallback:buffer` | 2min |
| Firebase Web | `firebase-web` | 3 | 60s | (shared with mobile) | 2min |
| Telegram | `telegram` | 3 | 120s | `telegram:pending` | 30s |
| Webhook | `webhook` | 3 | 60s | `webhook:pending` | 30s |
| Alerter | — | — | — | `alerter:pending` | 30s |

## Transition Logging Pattern

All components use the same pattern (established in log persist):
- **healthy → unhealthy**: one `Warn` log on transition
- **unhealthy → healthy**: one `Info` log on recovery
- No repeated error logging during downtime

## Files to modify

1. `internal/shared/infrastructure/firebase/firebase.go` — add breaker fields
2. `internal/shared/infrastructure/firebase/fcm.go` — wrap sends with CB, add fallback buffer + flusher
3. `internal/shared/infrastructure/telegram/client.go` — add breaker field
4. `internal/shared/infrastructure/telegram/sender.go` — wrap with CB, add pending buffer + flusher
5. `internal/shared/infrastructure/errors/webhook.go` — async conversion, CB, buffer + flusher
6. `internal/shared/infrastructure/errors/alerter.go` — add pending buffer + re-enqueue loop
7. `internal/shared/infrastructure/sse/bridge.go` — reconnect loop with exponential backoff
8. `internal/shared/infrastructure/eventbus/redis_streams.go` — in-memory fallback with atomic flag

## No new files needed

All changes are modifications to existing files. The `circuitbreaker.Breaker` is used as-is without modifications.
