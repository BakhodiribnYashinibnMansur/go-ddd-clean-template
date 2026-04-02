# Async Retry for External Services via Asynq

## Problem

Firebase/Telegram xato bersa notification yo'qoladi. Hozir error log'ga yoziladi, lekin qayta urinilmaydi.

## Solution

Barcha tashqi service call'larni Asynq task sifatida bajarish. Xato bo'lsa Asynq avtomatik retry qiladi (5 ta urinish, exponential backoff). 5 ta urinishdan keyin ham xato bo'lsa Dead Letter Queue'ga tushadi.

## Architecture

```
Handler/Service
    │
    ├─ Synchronous (sign-in, get user) → to'g'ridan-to'g'ri
    │
    └─ Async (notification, telegram, webhook)
        → Asynq task queue
            → Worker → Firebase/Telegram
            → Xato → retry (10s, 100s, 1000s, 10000s, 100000s)
            → 5x xato → Dead Letter Queue
```

## Task Types

| Task Type | Service | Payload |
|-----------|---------|---------|
| `task:send_fcm` | Firebase FCM | token, title, body, data, fcm_type |
| `task:send_fcm_multi` | Firebase multicast | tokens[], title, body, data, fcm_type |
| `task:send_telegram` | Telegram Bot | message_type, text |
| `task:send_webhook` | Generic HTTP | url, method, headers, body |

## Retry Policy

- MaxRetry: 5
- Backoff: exponential (Asynq default: `retry_count^4` seconds)
- Timeout per attempt: 30 seconds
- Dead Letter Queue: `asynq:dead` (Redis key)

## File Structure

```
internal/shared/infrastructure/asynq/
├── client.go              (mavjud)
├── worker.go              (mavjud)
├── tasks/
│   ├── types.go           ← task type constants + common options
│   ├── fcm.go             ← FCM send + multicast task handler
│   ├── telegram.go        ← Telegram message task handler
│   └── webhook.go         ← Generic webhook task handler
```

## Components

### 1. Task Types (`tasks/types.go`)

```go
const (
    TypeSendFCM      = "task:send_fcm"
    TypeSendFCMMulti = "task:send_fcm_multi"
    TypeSendTelegram = "task:send_telegram"
    TypeSendWebhook  = "task:send_webhook"
)

func DefaultRetryOpts() []asynq.Option {
    return []asynq.Option{
        asynq.MaxRetry(5),
        asynq.Timeout(30 * time.Second),
        asynq.Queue("external"),
    }
}
```

### 2. FCM Task (`tasks/fcm.go`)

```go
type FCMPayload struct {
    Token   string            `json:"token"`
    Title   string            `json:"title"`
    Body    string            `json:"body"`
    Data    map[string]string `json:"data,omitempty"`
    FCMType string            `json:"fcm_type"`
}

type FCMHandler struct {
    firebase *firebase.Firebase
}

func (h *FCMHandler) HandleSendFCM(ctx context.Context, t *asynq.Task) error {
    var p FCMPayload
    json.Unmarshal(t.Payload(), &p)
    return h.firebase.SendNotification(ctx, p.Token, p.FCMType,
        firebase.Content{Title: p.Title, Body: p.Body}, p.Data)
}
```

### 3. Telegram Task (`tasks/telegram.go`)

```go
type TelegramPayload struct {
    MessageType string `json:"message_type"`
    Text        string `json:"text"`
}

type TelegramHandler struct {
    client *telegram.Client
}

func (h *TelegramHandler) HandleSendTelegram(ctx context.Context, t *asynq.Task) error {
    var p TelegramPayload
    json.Unmarshal(t.Payload(), &p)
    return h.client.SendMessage(telegram.MessageType(p.MessageType), p.Text)
}
```

### 4. Worker Registration (`worker.go` update)

```go
mux.HandleFunc(tasks.TypeSendFCM, fcmHandler.HandleSendFCM)
mux.HandleFunc(tasks.TypeSendFCMMulti, fcmHandler.HandleSendFCMMulti)
mux.HandleFunc(tasks.TypeSendTelegram, telegramHandler.HandleSendTelegram)
mux.HandleFunc(tasks.TypeSendWebhook, webhookHandler.HandleSendWebhook)
```

### 5. Usage Change

```go
// BEFORE (sync, no retry):
f.MobileClient.Send(ctx, notification)

// AFTER (async, 5x retry):
asynqClient.EnqueueTask(ctx, tasks.TypeSendFCM, tasks.FCMPayload{
    Token: token, Title: "Yangi xabar", Body: "...", FCMType: "CLIENT",
}, tasks.DefaultRetryOpts()...)
```

## Error Integration

Worker ichida xato bo'lsa `HandleFirebaseError`/`HandleTelegramError` chaqiriladi — error metrics va logging ishlaydi. Asynq o'zi retry schedule'ni boshqaradi.

## Verification

1. Firebase token noto'g'ri → 5 retry → DLQ
2. Firebase timeout → 5 retry → keyingi urinishda ishlasa → success
3. Telegram rate limit → retry delay orqali o'tadi
4. `asynq stats` command orqali DLQ monitoring
