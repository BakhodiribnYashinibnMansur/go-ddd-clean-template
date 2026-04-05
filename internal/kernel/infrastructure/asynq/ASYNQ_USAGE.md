# Asynq Background Job Queue - Usage Examples

## Overview
Asynq - bu Redis-based background job queue tizimi. U email yuborish, rasm qayta ishlash va boshqa og'ir operatsiyalarni orqa fonda bajarish imkonini beradi.

## Architecture

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Client    │ ──────> │    Redis    │ <────── │   Worker    │
│ (Enqueue)   │         │   (Queue)   │         │  (Process)  │
└─────────────┘         └─────────────┘         └─────────────┘
```

## Configuration

`.env` faylda quyidagi sozlamalarni qo'shing:

```bash
# Asynq - Background Job Queue
ASYNQ_ADDR=localhost:6379
ASYNQ_PASSWORD=
ASYNQ_DB=0
ASYNQ_CONCURRENCY=10
ASYNQ_MAX_RETRY=3
ASYNQ_WORKER_ENABLED=true
```

## Usage Examples

### 1. Email Yuborish (Welcome Email)

```go
package main

import (
    "context"
    "gct/pkg/queue"
)

func SendWelcomeEmail(ctx context.Context, client *queue.Client, userEmail string) error {
    payload := queue.EmailPayload{
        To:      userEmail,
        Subject: "Welcome to Our Platform!",
        Body:    "Thank you for signing up...",
        Data: map[string]string{
            "username": "john_doe",
        },
    }

    // Enqueue task
    _, err := client.EnqueueEmail(ctx, queue.TypeEmailWelcome, payload)
    return err
}
```

### 2. Rasm Qayta Ishlash (Image Resize)

```go
func ResizeUserAvatar(ctx context.Context, client *queue.Client, imagePath string) error {
    payload := queue.ImagePayload{
        SourcePath: imagePath,
        TargetPath: "/uploads/avatars/resized_" + imagePath,
        Width:      200,
        Height:     200,
        Quality:    85,
    }

    // Enqueue with custom options
    opts := queue.TaskOptions{
        Queue:    queue.QueueLow,
        MaxRetry: 5,
        Timeout:  time.Minute * 5,
    }

    _, err := client.EnqueueImage(ctx, queue.TypeImageResize, payload, opts)
    return err
}
```

### 3. Push Notification Yuborish

```go
func SendPushNotification(ctx context.Context, client *queue.Client, userID string) error {
    payload := queue.NotificationPayload{
        UserID:  userID,
        Title:   "New Message",
        Message: "You have a new message from admin",
        Data: map[string]string{
            "type": "message",
            "from": "admin",
        },
    }

    _, err := client.EnqueueNotification(ctx, queue.TypePushNotification, payload)
    return err
}
```

### 4. Kechiktirilgan Task (Delayed Task)

```go
// 1 soatdan keyin email yuborish
func SendDelayedEmail(ctx context.Context, client *queue.Client) error {
    payload := queue.EmailPayload{
        To:      "user@example.com",
        Subject: "Reminder",
        Body:    "This is your reminder...",
    }

    // 1 soat kechikish bilan
    _, err := client.EnqueueTaskIn(
        ctx,
        queue.TypeEmailNotification,
        payload,
        time.Hour,
    )
    return err
}
```

### 5. Ma'lum Vaqtda Bajarilishi Kerak Bo'lgan Task

```go
// Ertaga soat 9:00 da email yuborish
func ScheduleEmail(ctx context.Context, client *queue.Client) error {
    payload := queue.EmailPayload{
        To:      "user@example.com",
        Subject: "Daily Report",
        Body:    "Your daily report...",
    }

    // Ertaga soat 9:00
    tomorrow9AM := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour).Add(9 * time.Hour)

    _, err := client.EnqueueTaskAt(
        ctx,
        queue.TypeEmailNotification,
        payload,
        tomorrow9AM,
    )
    return err
}
```

### 6. Controller dan Foydalanish

```go
package v1

import (
    "net/http"
    "gct/pkg/queue"
    "github.com/gin-gonic/gin"
)

type UserController struct {
    queueClient *queue.Client
}

func (uc *UserController) Register(c *gin.Context) {
    // ... user registration logic ...

    // Send welcome email in background
    payload := queue.EmailPayload{
        To:      user.Email,
        Subject: "Welcome!",
        Body:    "Thank you for registering...",
    }

    _, err := uc.queueClient.EnqueueEmail(c.Request.Context(), queue.TypeEmailWelcome, payload)
    if err != nil {
        // Log error but don't fail the request
        log.Error("failed to enqueue welcome email", err)
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}
```

## Queue Priorities

Asynq 3 ta queue priorityga ega:

1. **critical** (priority: 6) - Muhim tasklar (email verification, password reset)
2. **default** (priority: 3) - Oddiy tasklar (notifications)
3. **low** (priority: 1) - Kam muhim tasklar (image processing, cleanup)

## Monitoring

Asynq Web UI ni ishlatish mumkin:

```bash
# Install asynq CLI
go install github.com/hibiken/asynq/tools/asynq@latest

# Start web UI
asynq dash --redis-addr=localhost:6379
```

Web UI: `http://localhost:8080`

## Best Practices

1. **Idempotent Tasks**: Tasklar bir necha marta bajarilsa ham xavfsiz bo'lishi kerak
2. **Error Handling**: Har doim error handling qo'shing
3. **Timeouts**: Uzoq davom etadigan tasklarga timeout qo'ying
4. **Retry Logic**: Retry strategiyasini to'g'ri sozlang
5. **Monitoring**: Tasklar holatini monitoring qiling

## Custom Task Handler Qo'shish

1. `pkg/queue/tasks.go` ga yangi task type qo'shing:
```go
const TypeCustomTask = "custom:task"
```

2. `pkg/queue/handlers.go` ga handler qo'shing:
```go
func (h *Handlers) HandleCustomTask(ctx context.Context, task *asynq.Task) error {
    // Your logic here
    return nil
}
```

3. `internal/app/app.go` da handler ni register qiling:
```go
asynqWorker.RegisterHandler(queue.TypeCustomTask, handlers.HandleCustomTask)
```
