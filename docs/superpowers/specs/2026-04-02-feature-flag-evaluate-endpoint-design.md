# Feature Flag Evaluate Endpoint Design

## Overview

Feature flag tizimiga evaluate endpoint qo'shish — klientlar (web frontend, mobile) va backend service'lar feature flaglarni `userAttrs` (jumladan `platform`) asosida baholay olishi uchun. Hozirgi condition tizimi (rule group + condition) orqali platform filtering amalga oshiriladi — yangi domain tushuncha kiritilmaydi.

## Qarorlar

- **Yondashuv A (condition-based platform filtering)** tanlandi — `platform` oddiy `userAttrs` attribute sifatida keladi, admin rule group condition orqali boshqaradi
- **CQRS query handler** orqali — mavjud arxitektura pattern'ga mos
- **Authenticated endpoint** — mavjud auth middleware bilan himoyalangan
- **Bitta va batch** evaluate — ikkalasi ham POST (user_attrs body'da)

## API Endpoint'lar

### Bitta flag evaluate

```
POST /api/v1/feature-flags/evaluate
Authorization: Bearer <token>

Request:
{
  "key": "new_dashboard",
  "user_attrs": {
    "platform": "web",
    "user_id": "abc-123",
    "region": "uz"
  }
}

Response 200:
{
  "key": "new_dashboard",
  "value": "true",
  "flag_type": "bool"
}

Response 404 (flag topilmasa):
{
  "error": "feature flag not found"
}
```

### Batch evaluate

```
POST /api/v1/feature-flags/evaluate/batch
Authorization: Bearer <token>

Request:
{
  "keys": ["new_dashboard", "dark_mode", "beta_checkout"],
  "user_attrs": {
    "platform": "web",
    "user_id": "abc-123"
  }
}

Response 200:
{
  "flags": {
    "new_dashboard": {"value": "true", "flag_type": "bool"},
    "dark_mode": {"value": "dark", "flag_type": "string"},
    "beta_checkout": {"value": "false", "flag_type": "bool"}
  }
}
```

Topilmagan flaglar response'dan chiqarib tashlanadi (xato qaytarilmaydi).

## Arxitektura

### Yangi fayllar

| Fayl | Vazifasi |
|------|----------|
| `application/query/evaluate.go` | `EvaluateQuery`, `EvaluateResult`, `EvaluateHandler` |
| `application/query/evaluate_batch.go` | `BatchEvaluateQuery`, `BatchEvaluateResult`, `BatchEvaluateHandler` |

### O'zgartiriladigan fayllar

| Fayl | O'zgarish |
|------|-----------|
| `infrastructure/cache/evaluator_cache.go` | `EvaluateFull()` metodi qo'shiladi — `value` + `flagType` qaytaradi |
| `interfaces/http/handler.go` | `Evaluate()` va `BatchEvaluate()` handler metodlari |
| `interfaces/http/routes.go` | Ikki yangi route |
| `interfaces/http/request.go` | `EvaluateRequest` va `BatchEvaluateRequest` struct'lar |
| `bc.go` | `EvaluateFlag` va `BatchEvaluateFlag` field'lar |

### CQRS Query Layer

```go
// query/evaluate.go
type EvaluateQuery struct {
    Key       string
    UserAttrs map[string]string
}

type EvaluateResult struct {
    Key      string
    Value    string
    FlagType string
}

type EvaluateHandler struct {
    evaluator *cache.CachedEvaluator
}

func (h *EvaluateHandler) Handle(ctx context.Context, q EvaluateQuery) (*EvaluateResult, error)
```

```go
// query/evaluate_batch.go
type BatchEvaluateQuery struct {
    Keys      []string
    UserAttrs map[string]string
}

type BatchEvaluateResult struct {
    Flags map[string]EvaluateResult
}

type BatchEvaluateHandler struct {
    evaluator *cache.CachedEvaluator
}

func (h *BatchEvaluateHandler) Handle(ctx context.Context, q BatchEvaluateQuery) (*BatchEvaluateResult, error)
```

### CachedEvaluator yangi metod

```go
// EvalResult holds the evaluated value and flag type.
type EvalResult struct {
    Value    string
    FlagType string
}

func (ce *CachedEvaluator) EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *EvalResult
```

Flag topilmasa `nil` qaytaradi. Ichida `getFlag()` chaqiriladi, keyin `ff.Evaluate(userAttrs)` va `ff.FlagType()`.

### HTTP Handler

```go
// request.go
type EvaluateRequest struct {
    Key       string            `json:"key" binding:"required"`
    UserAttrs map[string]string `json:"user_attrs"`
}

type BatchEvaluateRequest struct {
    Keys      []string          `json:"keys" binding:"required,min=1"`
    UserAttrs map[string]string `json:"user_attrs"`
}
```

### Route'lar

```go
g.POST("/evaluate", h.Evaluate)
g.POST("/evaluate/batch", h.BatchEvaluate)
```

### BoundedContext

```go
type BoundedContext struct {
    // ... mavjud fieldlar
    EvaluateFlag      *query.EvaluateHandler
    BatchEvaluateFlag *query.BatchEvaluateHandler
}
```

## Platform Filtering misoli

Admin `new_dashboard` (bool, default: "false") flag'ga rule group yaratadi:

```json
POST /api/v1/feature-flags/:id/rule-groups
{
  "name": "Web only",
  "variation": "true",
  "priority": 1,
  "conditions": [
    {"attribute": "platform", "operator": "eq", "value": "web"}
  ]
}
```

**Web klient:**
```json
POST /api/v1/feature-flags/evaluate
{"key": "new_dashboard", "user_attrs": {"platform": "web", "user_id": "u1"}}
→ {"key": "new_dashboard", "value": "true", "flag_type": "bool"}
```

**Mobile klient:**
```json
POST /api/v1/feature-flags/evaluate
{"key": "new_dashboard", "user_attrs": {"platform": "mobile", "user_id": "u1"}}
→ {"key": "new_dashboard", "value": "false", "flag_type": "bool"}
```

## Test strategiyasi

- **Unit test**: `EvaluateHandler` va `BatchEvaluateHandler` — mock evaluator bilan
- **Unit test**: `CachedEvaluator.EvaluateFull()` — mavjud cache test pattern bilan
- **HTTP test**: Evaluate va BatchEvaluate endpoint'lar — request parsing, response format, 404 holat
