package command_test

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/iam/session/application/command"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockEventBus struct {
	publishedEvents []shareddomain.DomainEvent
	publishErr      error
}

func (m *mockEventBus) Publish(_ context.Context, events ...shareddomain.DomainEvent) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                    {}
func (m *mockLogger) Debugf(template string, args ...any)                  {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Info(args ...any)                                     {}
func (m *mockLogger) Infof(template string, args ...any)                   {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)               {}
func (m *mockLogger) Warn(args ...any)                                     {}
func (m *mockLogger) Warnf(template string, args ...any)                   {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)               {}
func (m *mockLogger) Error(args ...any)                                    {}
func (m *mockLogger) Errorf(template string, args ...any)                  {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Fatal(args ...any)                                    {}
func (m *mockLogger) Fatalf(template string, args ...any)                  {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// ---------------------------------------------------------------------------
// RevokeSession tests
// ---------------------------------------------------------------------------

func TestRevokeSessionHandler_Handle_Success(t *testing.T) {
	eb := &mockEventBus{}
	handler := command.NewRevokeSessionHandler(eb, &mockLogger{})

	userID := uuid.New()
	sessionID := uuid.New()

	err := handler.Handle(context.Background(), command.RevokeSessionCommand{
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(eb.publishedEvents) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(eb.publishedEvents))
	}
	if eb.publishedEvents[0].EventName() != "session.revoke_requested" {
		t.Fatalf("expected session.revoke_requested, got %s", eb.publishedEvents[0].EventName())
	}
	if eb.publishedEvents[0].AggregateID() != userID {
		t.Fatalf("expected aggregate ID %s, got %s", userID, eb.publishedEvents[0].AggregateID())
	}
}

func TestRevokeSessionHandler_Handle_PublishError(t *testing.T) {
	publishErr := errors.New("publish failed")
	eb := &mockEventBus{publishErr: publishErr}
	handler := command.NewRevokeSessionHandler(eb, &mockLogger{})

	err := handler.Handle(context.Background(), command.RevokeSessionCommand{
		UserID:    uuid.New(),
		SessionID: uuid.New(),
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("expected publish error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// RevokeAllSessions tests
// ---------------------------------------------------------------------------

func TestRevokeAllSessionsHandler_Handle_Success(t *testing.T) {
	eb := &mockEventBus{}
	handler := command.NewRevokeAllSessionsHandler(eb, &mockLogger{})

	userID := uuid.New()

	err := handler.Handle(context.Background(), command.RevokeAllSessionsCommand{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(eb.publishedEvents) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(eb.publishedEvents))
	}
	if eb.publishedEvents[0].EventName() != "session.revoke_all_requested" {
		t.Fatalf("expected session.revoke_all_requested, got %s", eb.publishedEvents[0].EventName())
	}
	if eb.publishedEvents[0].AggregateID() != userID {
		t.Fatalf("expected aggregate ID %s, got %s", userID, eb.publishedEvents[0].AggregateID())
	}
}

func TestRevokeAllSessionsHandler_Handle_PublishError(t *testing.T) {
	publishErr := errors.New("bus down")
	eb := &mockEventBus{publishErr: publishErr}
	handler := command.NewRevokeAllSessionsHandler(eb, &mockLogger{})

	err := handler.Handle(context.Background(), command.RevokeAllSessionsCommand{
		UserID: uuid.New(),
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("expected publish error, got %v", err)
	}
}
