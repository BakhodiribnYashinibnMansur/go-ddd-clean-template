package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gct/internal/platform/infrastructure/asynq/tasks"

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

	payload, _ := json.Marshal(tasks.TelegramPayload{MessageType: "error", Text: "Something broke"})

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

	payload, _ := json.Marshal(tasks.TelegramPayload{MessageType: "info", Text: "test"})

	err := h.HandleSendTelegram(context.Background(), asynq.NewTask(tasks.TypeSendTelegram, payload))
	if err == nil {
		t.Fatal("expected error")
	}
}
