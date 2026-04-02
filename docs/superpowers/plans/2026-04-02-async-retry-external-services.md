# Async Retry for External Services Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Firebase, Telegram va boshqa tashqi service call'larni Asynq orqali async retry bilan bajarish — xato bo'lsa 5 marta qayta urinish, DLQ ga tushish.

**Architecture:** Tashqi service call'lar to'g'ridan-to'g'ri emas, Asynq task queue orqali bajariladi. Worker xatoni olsa Asynq avtomatik exponential backoff bilan retry qiladi. 5 ta xatodan keyin Dead Letter Queue'ga tushadi.

**Tech Stack:** Asynq (mavjud), Redis (mavjud), Firebase FCM (mavjud), Telegram Bot (mavjud)

---

## File Structure

```
internal/shared/infrastructure/asynq/
├── client.go                  (mavjud — o'zgarmaydi)
├── worker.go                  (mavjud — RegisterHandler chaqiruvlar qo'shiladi)
├── tasks/
│   ├── types.go               ← YANGI: task type constants + DefaultRetryOpts
│   ├── fcm.go                 ← YANGI: FCM send + multicast task handler
│   ├── fcm_test.go            ← YANGI: FCM handler tests
│   ├── telegram.go            ← YANGI: Telegram message task handler
│   └── telegram_test.go       ← YANGI: Telegram handler tests
config/
├── asynq.go                   (mavjud — "external" queue qo'shiladi)
internal/shared/infrastructure/firebase/
├── fcm.go                     (mavjud — o'zgarmaydi, worker ichidan chaqiriladi)
internal/shared/infrastructure/telegram/
├── sender.go                  (mavjud — o'zgarmaydi, worker ichidan chaqiriladi)
internal/app/
├── app.go                     (mavjud — worker handler registration qo'shiladi)
```

---

### Task 1: Task Types va Constants

**Files:**
- Create: `internal/shared/infrastructure/asynq/tasks/types.go`

- [ ] **Step 1: Create task types file**

```go
// internal/shared/infrastructure/asynq/tasks/types.go
package tasks

import (
	"time"

	"github.com/hibiken/asynq"
)

// Task type constants.
const (
	TypeSendFCM      = "task:send_fcm"
	TypeSendFCMMulti = "task:send_fcm_multi"
	TypeSendTelegram = "task:send_telegram"
)

// DefaultRetryOpts returns standard retry options for external service tasks.
// 5 retries with exponential backoff (Asynq default: retry_count^4 seconds).
// Tasks go to "external" queue and timeout after 30 seconds per attempt.
func DefaultRetryOpts() []asynq.Option {
	return []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.Queue("external"),
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/shared/infrastructure/asynq/...`
Expected: no errors

- [ ] **Step 3: Add "external" queue to default config**

Modify: `config/asynq.go:32-36`

```go
// Default queue priorities
return map[string]int{
	"critical": 6, // Highest priority
	"default":  3, // Medium priority
	"external": 2, // External service retries
	"low":      1, // Lowest priority
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/shared/infrastructure/asynq/tasks/types.go config/asynq.go
git commit -m "feat(asynq): add external service task types and retry options"
```

---

### Task 2: FCM Task Handler

**Files:**
- Create: `internal/shared/infrastructure/asynq/tasks/fcm.go`
- Create: `internal/shared/infrastructure/asynq/tasks/fcm_test.go`

- [ ] **Step 1: Write FCM task handler test**

```go
// internal/shared/infrastructure/asynq/tasks/fcm_test.go
package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/asynq/tasks"

	"github.com/hibiken/asynq"
)

type mockFirebase struct {
	sendErr      error
	multiSendErr error
	sendCalled   bool
	multiCalled  bool
}

func (m *mockFirebase) Send(ctx context.Context, token, fcmType, title, body string, data map[string]string) error {
	m.sendCalled = true
	return m.sendErr
}

func (m *mockFirebase) SendMulti(ctx context.Context, tokens []string, fcmType, title, body string, data map[string]string) error {
	m.multiCalled = true
	return m.multiSendErr
}

func TestFCMHandler_HandleSendFCM_Success(t *testing.T) {
	fb := &mockFirebase{}
	h := tasks.NewFCMHandler(fb)

	payload, _ := json.Marshal(tasks.FCMPayload{
		Token:   "test-token",
		Title:   "Test",
		Body:    "Hello",
		FCMType: "CLIENT",
	})

	err := h.HandleSendFCM(context.Background(), asynq.NewTask(tasks.TypeSendFCM, payload))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !fb.sendCalled {
		t.Fatal("expected Send to be called")
	}
}

func TestFCMHandler_HandleSendFCM_Error(t *testing.T) {
	fb := &mockFirebase{sendErr: errors.New("firebase unavailable")}
	h := tasks.NewFCMHandler(fb)

	payload, _ := json.Marshal(tasks.FCMPayload{
		Token:   "test-token",
		Title:   "Test",
		Body:    "Hello",
		FCMType: "CLIENT",
	})

	err := h.HandleSendFCM(context.Background(), asynq.NewTask(tasks.TypeSendFCM, payload))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFCMHandler_HandleSendFCMMulti_Success(t *testing.T) {
	fb := &mockFirebase{}
	h := tasks.NewFCMHandler(fb)

	payload, _ := json.Marshal(tasks.FCMMultiPayload{
		Tokens:  []string{"token1", "token2"},
		Title:   "Test",
		Body:    "Hello",
		FCMType: "CLIENT",
	})

	err := h.HandleSendFCMMulti(context.Background(), asynq.NewTask(tasks.TypeSendFCMMulti, payload))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !fb.multiCalled {
		t.Fatal("expected SendMulti to be called")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/asynq/tasks/... -v`
Expected: FAIL — `tasks.FCMPayload` undefined

- [ ] **Step 3: Write FCM task handler implementation**

```go
// internal/shared/infrastructure/asynq/tasks/fcm.go
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// FCMSender abstracts Firebase notification sending for testability.
type FCMSender interface {
	Send(ctx context.Context, token, fcmType, title, body string, data map[string]string) error
	SendMulti(ctx context.Context, tokens []string, fcmType, title, body string, data map[string]string) error
}

// FCMPayload is the task payload for a single FCM notification.
type FCMPayload struct {
	Token   string            `json:"token"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
	FCMType string            `json:"fcm_type"`
}

// FCMMultiPayload is the task payload for a multicast FCM notification.
type FCMMultiPayload struct {
	Tokens  []string          `json:"tokens"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
	FCMType string            `json:"fcm_type"`
}

// FCMHandler processes FCM notification tasks.
type FCMHandler struct {
	sender FCMSender
}

// NewFCMHandler creates a new FCM task handler.
func NewFCMHandler(sender FCMSender) *FCMHandler {
	return &FCMHandler{sender: sender}
}

// HandleSendFCM sends a single FCM notification. Returns error to trigger Asynq retry.
func (h *FCMHandler) HandleSendFCM(ctx context.Context, t *asynq.Task) error {
	var p FCMPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal FCM payload: %w", err)
	}
	return h.sender.Send(ctx, p.Token, p.FCMType, p.Title, p.Body, p.Data)
}

// HandleSendFCMMulti sends a multicast FCM notification. Returns error to trigger Asynq retry.
func (h *FCMHandler) HandleSendFCMMulti(ctx context.Context, t *asynq.Task) error {
	var p FCMMultiPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal FCM multi payload: %w", err)
	}
	return h.sender.SendMulti(ctx, p.Tokens, p.FCMType, p.Title, p.Body, p.Data)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/asynq/tasks/... -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/asynq/tasks/fcm.go internal/shared/infrastructure/asynq/tasks/fcm_test.go
git commit -m "feat(asynq): add FCM notification task handler with tests"
```

---

### Task 3: Telegram Task Handler

**Files:**
- Create: `internal/shared/infrastructure/asynq/tasks/telegram.go`
- Create: `internal/shared/infrastructure/asynq/tasks/telegram_test.go`

- [ ] **Step 1: Write Telegram task handler test**

```go
// internal/shared/infrastructure/asynq/tasks/telegram_test.go
package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/asynq/tasks"

	"github.com/hibiken/asynq"
)

type mockTelegram struct {
	sendErr    error
	sendCalled bool
	lastType   string
	lastText   string
}

func (m *mockTelegram) Send(msgType, text string) error {
	m.sendCalled = true
	m.lastType = msgType
	m.lastText = text
	return m.sendErr
}

func TestTelegramHandler_Success(t *testing.T) {
	tg := &mockTelegram{}
	h := tasks.NewTelegramHandler(tg)

	payload, _ := json.Marshal(tasks.TelegramPayload{
		MessageType: "error",
		Text:        "Something broke",
	})

	err := h.HandleSendTelegram(context.Background(), asynq.NewTask(tasks.TypeSendTelegram, payload))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tg.sendCalled {
		t.Fatal("expected Send to be called")
	}
	if tg.lastType != "error" {
		t.Fatalf("expected type 'error', got %q", tg.lastType)
	}
}

func TestTelegramHandler_Error(t *testing.T) {
	tg := &mockTelegram{sendErr: errors.New("telegram unavailable")}
	h := tasks.NewTelegramHandler(tg)

	payload, _ := json.Marshal(tasks.TelegramPayload{
		MessageType: "info",
		Text:        "test",
	})

	err := h.HandleSendTelegram(context.Background(), asynq.NewTask(tasks.TypeSendTelegram, payload))
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/asynq/tasks/... -run TestTelegram -v`
Expected: FAIL

- [ ] **Step 3: Write Telegram task handler implementation**

```go
// internal/shared/infrastructure/asynq/tasks/telegram.go
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// TelegramSender abstracts Telegram message sending for testability.
type TelegramSender interface {
	Send(msgType, text string) error
}

// TelegramPayload is the task payload for a Telegram message.
type TelegramPayload struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

// TelegramHandler processes Telegram message tasks.
type TelegramHandler struct {
	sender TelegramSender
}

// NewTelegramHandler creates a new Telegram task handler.
func NewTelegramHandler(sender TelegramSender) *TelegramHandler {
	return &TelegramHandler{sender: sender}
}

// HandleSendTelegram sends a Telegram message. Returns error to trigger Asynq retry.
func (h *TelegramHandler) HandleSendTelegram(ctx context.Context, t *asynq.Task) error {
	var p TelegramPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal Telegram payload: %w", err)
	}
	return h.sender.Send(p.MessageType, p.Text)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/asynq/tasks/... -v`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/asynq/tasks/telegram.go internal/shared/infrastructure/asynq/tasks/telegram_test.go
git commit -m "feat(asynq): add Telegram message task handler with tests"
```

---

### Task 4: Firebase FCMSender Adapter

**Files:**
- Create: `internal/shared/infrastructure/firebase/adapter.go`

Firebase'ning mavjud `SendNotification`/`SendMultiNotification` method'larini `FCMSender` interface'ga moslashtirish kerak.

- [ ] **Step 1: Create Firebase adapter**

```go
// internal/shared/infrastructure/firebase/adapter.go
package firebase

import "context"

// TaskAdapter adapts Firebase to the tasks.FCMSender interface.
type TaskAdapter struct {
	fb *Firebase
}

// NewTaskAdapter creates a new Firebase task adapter.
func NewTaskAdapter(fb *Firebase) *TaskAdapter {
	return &TaskAdapter{fb: fb}
}

// Send sends a single FCM notification.
func (a *TaskAdapter) Send(ctx context.Context, token, fcmType, title, body string, data map[string]string) error {
	return a.fb.SendNotification(ctx, token, fcmType, Content{Title: title, Body: body}, data)
}

// SendMulti sends a multicast FCM notification.
func (a *TaskAdapter) SendMulti(ctx context.Context, tokens []string, fcmType, title, body string, data map[string]string) error {
	return a.fb.SendMultiNotification(ctx, tokens, fcmType, Content{Title: title, Body: body}, data)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/shared/infrastructure/firebase/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/shared/infrastructure/firebase/adapter.go
git commit -m "feat(firebase): add TaskAdapter for Asynq FCMSender interface"
```

---

### Task 5: Telegram TaskAdapter

**Files:**
- Create: `internal/shared/infrastructure/telegram/adapter.go`

- [ ] **Step 1: Create Telegram adapter**

```go
// internal/shared/infrastructure/telegram/adapter.go
package telegram

// TaskAdapter adapts telegram.Client to the tasks.TelegramSender interface.
type TaskAdapter struct {
	client *Client
}

// NewTaskAdapter creates a new Telegram task adapter.
func NewTaskAdapter(client *Client) *TaskAdapter {
	return &TaskAdapter{client: client}
}

// Send sends a Telegram message via the underlying client.
func (a *TaskAdapter) Send(msgType, text string) error {
	return a.client.SendMessage(MessageType(msgType), text)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/shared/infrastructure/telegram/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/shared/infrastructure/telegram/adapter.go
git commit -m "feat(telegram): add TaskAdapter for Asynq TelegramSender interface"
```

---

### Task 6: Worker Registration

**Files:**
- Modify: `internal/shared/infrastructure/asynq/worker.go`
- Modify: `internal/app/app.go` (worker handler registration)

- [ ] **Step 1: Add RegisterExternalHandlers method to Worker**

Add to `internal/shared/infrastructure/asynq/worker.go`:

```go
// RegisterExternalHandlers registers all external service task handlers.
func (w *Worker) RegisterExternalHandlers(fcm *tasks.FCMHandler, tg *tasks.TelegramHandler) {
	if fcm != nil {
		w.RegisterHandler(tasks.TypeSendFCM, fcm.HandleSendFCM)
		w.RegisterHandler(tasks.TypeSendFCMMulti, fcm.HandleSendFCMMulti)
		w.log.Info("📤 Registered FCM task handlers")
	}
	if tg != nil {
		w.RegisterHandler(tasks.TypeSendTelegram, tg.HandleSendTelegram)
		w.log.Info("📤 Registered Telegram task handlers")
	}
}
```

Add import: `"gct/internal/shared/infrastructure/asynq/tasks"`

- [ ] **Step 2: Wire handlers in app.go**

Find where the Asynq worker is initialized in `internal/app/app.go` and add handler registration after worker creation. The exact location depends on current app.go structure — look for `asynq.NewWorker` or `worker.Start()` and add before `Start()`:

```go
// Register external service task handlers
var fcmHandler *tasks.FCMHandler
if app.firebase != nil {
	fcmHandler = tasks.NewFCMHandler(firebase.NewTaskAdapter(app.firebase))
}

var tgHandler *tasks.TelegramHandler
if app.telegram != nil {
	tgHandler = tasks.NewTelegramHandler(telegram.NewTaskAdapter(app.telegram))
}

app.worker.RegisterExternalHandlers(fcmHandler, tgHandler)
```

Add imports:
```go
"gct/internal/shared/infrastructure/asynq/tasks"
"gct/internal/shared/infrastructure/firebase"
"gct/internal/shared/infrastructure/telegram"
```

- [ ] **Step 3: Verify full project compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/shared/infrastructure/asynq/worker.go internal/app/app.go
git commit -m "feat(asynq): register external service task handlers in worker"
```

---

### Task 7: Full Build and Test Verification

- [ ] **Step 1: Build entire project**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: no errors

- [ ] **Step 2: Run task handler tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/asynq/... -v`
Expected: all tests PASS

- [ ] **Step 3: Run full test suite**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test $(go list ./... | grep -v "test/e2e/flows/user/client") 2>&1 | grep FAIL`
Expected: no FAIL lines

- [ ] **Step 4: Run go vet**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go vet ./...`
Expected: no issues
