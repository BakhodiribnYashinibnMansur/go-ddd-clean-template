# Error Observability — 5 Feature Spec

## 1. Request ID Correlation

Bitta `request_id` (UUID) har joyda bir xil.

- **Middleware**: `X-Request-ID` header'dan oladi yoki yangi UUID yaratadi → context'ga qo'yadi
- **Response header**: `X-Request-ID` response'da qaytaradi
- **Error response**: JSON'da `requestId` context'dan olinadi (hardcoded UUID.New() emas)
- **Logger**: har bir log entry'da `request_id` avtomatik
- **OTel span**: `request_id` attribute sifatida Jaeger'da ko'rinadi

## 2. Error Alerting (Telegram)

CRITICAL/HIGH severity error bo'lsa Telegram'ga avtomatik xabar.

- **Reporter implementation**: mavjud `Reporter` interface'ni implement qilish (errors/logging.go)
- **Alert content**: error code, request_id, severity, timestamp, path
- **Debounce**: bir xil error code uchun 1 minutda max 1 ta alert (spam oldini olish)
- **Asynq orqali**: alert Telegram task sifatida yuboriladi (retry bilan)
- **Configurable**: qaysi severity alert yuborishi config'dan

## 3. Circuit Breaker

Tashqi service ishlamasa avtomatik o'chirish.

- **Pattern**: CLOSED → OPEN → HALF-OPEN → CLOSED
- **Config per service**: failure threshold (5), timeout (30s), half-open max requests (1)
- **Services**: Firebase, Telegram (Asynq worker ichida)
- **Alert integration**: circuit OPEN bo'lganda Telegram alert
- **Metrics**: circuit state o'zgarishi log'ga yoziladi

## 4. Error Rate Monitoring

Error rate limitdan oshsa alert.

- **Sliding window**: oxirgi 1 minutdagi error count per error code
- **Threshold**: configurable (default: 10 errors/min per code)
- **Alert**: threshold oshganda Telegram alert (debounced)
- **Metrics hook**: mavjud `ErrorHookManager` ga rate monitoring hook qo'shiladi

## 5. Health Check Endpoints

`/health`, `/ready` endpoint'lar.

- **`GET /health`** — server alive (always 200 if process running)
- **`GET /ready`** — barcha dependency'lar tekshiriladi:
  - PostgreSQL: `SELECT 1`
  - Redis: `PING`
  - MinIO: `BucketExists`
  - Asynq: Redis connection check
- **Response format**: `{"status": "ok/degraded/unhealthy", "checks": {...}, "uptime": "..."}`
- **Circuit breaker state**: health response'da ko'rinadi
