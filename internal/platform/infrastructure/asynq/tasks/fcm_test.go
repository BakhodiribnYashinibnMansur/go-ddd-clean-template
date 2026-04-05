package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gct/internal/platform/infrastructure/asynq/tasks"

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
		Token: "test-token", Title: "Test", Body: "Hello", FCMType: "CLIENT",
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
		Token: "test-token", Title: "Test", Body: "Hello", FCMType: "CLIENT",
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
		Tokens: []string{"token1", "token2"}, Title: "Test", Body: "Hello", FCMType: "CLIENT",
	})

	err := h.HandleSendFCMMulti(context.Background(), asynq.NewTask(tasks.TypeSendFCMMulti, payload))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !fb.multiCalled {
		t.Fatal("expected SendMulti to be called")
	}
}
