# Asynq Background Job Queue Integration

## Overview

Asynq - bu Redis-based background job queue tizimi. Bu loyihaga email yuborish, rasm qayta ishlash va boshqa og'ir operatsiyalarni orqa fonda bajarish imkonini beradi.

## Features

✅ **Redis-based Queue** - Mavjud Redis infrastrukturasidan foydalanadi  
✅ **Priority Queues** - 3 darajali prioritet (critical, default, low)  
✅ **Retry Logic** - Avtomatik retry mexanizmi  
✅ **Graceful Shutdown** - Xavfsiz to'xtatish  
✅ **Task Scheduling** - Kechiktirilgan va rejalashtirilgan tasklar  
✅ **Monitoring Ready** - Asynq Web UI bilan monitoring  

## Architecture

```
┌──────────────┐         ┌─────────────┐         ┌──────────────┐
│   Client     │ ──────> │    Redis    │ <────── │   Worker     │
│  (Enqueue)   │         │   (Queue)   │         │  (Process)   │
└──────────────┘         └─────────────┘         └──────────────┘
```

## Quick Start

### 1. Configuration

`.env` faylga qo'shing:

```bash
# Asynq - Background Job Queue
ASYNQ_ADDR=localhost:6379
ASYNQ_PASSWORD=
ASYNQ_DB=0
ASYNQ_CONCURRENCY=10
ASYNQ_MAX_RETRY=3
ASYNQ_WORKER_ENABLED=true
```

### 2. Available Task Types

- **Email Tasks**: `email:welcome`, `email:verification`, `email:password_reset`
- **Image Processing**: `image:resize`, `image:optimize`, `image:thumbnail`
- **Notifications**: `notification:push`, `notification:sms`
- **Reports**: `report:generate`, `report:export`
- **Cleanup**: `cleanup:old_sessions`, `cleanup:temp_files`

### 3. Usage Examples

#### From Controller/UseCase

```go
import "gct/pkg/asynq"

// Email yuborish
payload := asynq.EmailPayload{
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    "Thank you for registering...",
}

_, err := uc.AsynqClient.EnqueueEmail(ctx, asynq.TypeEmailWelcome, payload)
```

#### Kechiktirilgan Task

```go
// 1 soatdan keyin bajarish
_, err := uc.AsynqClient.EnqueueTaskIn(
    ctx,
    asynq.TypeEmailNotification,
    payload,
    time.Hour,
)
```

#### Custom Options

```go
opts := asynq.TaskOptions{
    Queue:    asynq.QueueCritical,  // High priority
    MaxRetry: 5,
    Timeout:  5 * time.Minute,
}

_, err := uc.AsynqClient.EnqueueImage(ctx, asynq.TypeImageResize, payload, opts)
```

## Test Endpoints

Loyihada test uchun 3 ta endpoint mavjud:

### 1. Email Test
```bash
POST /api/v1/asynq/email/test
Content-Type: application/json

{
  "to": "user@example.com",
  "subject": "Test Email",
  "body": "This is a test email",
  "data": {
    "username": "john_doe"
  }
}
```

### 2. Image Processing Test
```bash
POST /api/v1/asynq/image/process
Content-Type: application/json

{
  "source_path": "/uploads/original.jpg",
  "target_path": "/uploads/resized.jpg",
  "width": 200,
  "height": 200,
  "quality": 85
}
```

### 3. Notification Test
```bash
POST /api/v1/asynq/notification/test
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "New Message",
  "message": "You have a new message",
  "data": {
    "type": "message"
  }
}
```

## Monitoring

### Asynq Web UI

```bash
# Install CLI
go install github.com/hibiken/asynq/tools/asynq@latest

# Start dashboard
asynq dash --redis-addr=localhost:6379

# Open browser
open http://localhost:8080
```

### Features:
- Task queue monitoring
- Failed task inspection
- Retry management
- Performance metrics

## Adding Custom Tasks

### 1. Define Task Type

`pkg/asynq/tasks.go`:
```go
const TypeCustomTask = "custom:my_task"
```

### 2. Create Handler

`pkg/asynq/handlers.go`:
```go
func (h *Handlers) HandleCustomTask(ctx context.Context, task *asynq.Task) error {
    var payload MyPayload
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return err
    }
    
    // Your logic here
    
    return nil
}
```

### 3. Register Handler

`internal/app/app.go`:
```go
asynqWorker.RegisterHandler(asynq.TypeCustomTask, handlers.HandleCustomTask)
```

## Best Practices

1. **Idempotency** - Tasklar bir necha marta bajarilsa ham xavfsiz bo'lishi kerak
2. **Error Handling** - Har doim error handling qo'shing
3. **Timeouts** - Uzoq davom etadigan tasklarga timeout qo'ying
4. **Logging** - Har bir task boshlanishi va tugashini log qiling
5. **Monitoring** - Production da Asynq Web UI yoki Prometheus metrics ishlatish

## Production Considerations

- **Redis HA**: Production da Redis Sentinel yoki Cluster ishlatish tavsiya etiladi
- **Worker Scaling**: Yuqori load uchun worker concurrency ni oshiring
- **Queue Priorities**: Muhim tasklar uchun `critical` queue ishlatish
- **Retention**: Tugallangan tasklarni saqlash muddatini sozlang
- **Metrics**: Prometheus metrics ni yoqing va monitoring qiling

## Documentation

- [Full Usage Guide](./ASYNQ_USAGE.md)
- [Asynq Official Docs](https://github.com/hibiken/asynq)

## Files Created

```
pkg/asynq/
├── client.go       # Asynq client wrapper
├── worker.go       # Asynq worker
├── tasks.go        # Task type constants
├── handlers.go     # Task handlers
└── helpers.go      # Helper functions

internal/controller/restapi/v1/asynq/
├── controller.go   # Test endpoints controller
└── router.go       # Asynq routes

config/
└── asynq.go        # Asynq configuration

docs/
└── ASYNQ_USAGE.md  # Detailed usage guide
```

## Summary

Asynq tizimi muvaffaqiyatli integratsiya qilindi! Endi siz:

- ✅ Email yuborishni background da bajarasiz
- ✅ Rasm qayta ishlashni queue ga qo'yasiz
- ✅ Notificationlarni async yuborasiz
- ✅ Har qanday og'ir operatsiyani orqa fonda bajarasiz
- ✅ Task monitoring va retry mexanizmiga ega bo'lasiz
