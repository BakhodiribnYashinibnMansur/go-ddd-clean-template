package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNoopLogger_DoesNotPanic(t *testing.T) {
	var l NoopLogger

	uid := uuid.New()
	sid := uuid.New()

	l.Log(context.Background(), Entry{
		Event:           EventSignInSuccess,
		IntegrationName: "web",
		UserID:          &uid,
		SessionID:       &sid,
		IPAddress:       "127.0.0.1",
		UserAgent:       "test/1.0",
		Metadata:        map[string]any{"key": "value"},
	})

	// nil optional fields
	l.Log(context.Background(), Entry{
		Event: EventAPIKeyMismatch,
	})
}

func TestPgLogger_DropsWhenChannelFull(t *testing.T) {
	// Pass nil pool — we never flush so it won't be touched.
	pg := NewPgLogger(nil, noopLog{})

	// Do NOT call Start — channel is unbuffered consumer-side,
	// so it will fill up at channelSize.
	for i := 0; i < channelSize+100; i++ {
		pg.Log(context.Background(), Entry{Event: EventSignInFailed})
	}

	if dropped := pg.Dropped(); dropped == 0 {
		t.Fatal("expected dropped > 0 when channel is full")
	}
}

func TestEntry_NilUserAndSession(t *testing.T) {
	e := Entry{
		Event:     EventRefreshReuse,
		UserID:    nil,
		SessionID: nil,
		Metadata:  nil,
	}

	if e.UserID != nil {
		t.Fatal("expected nil UserID")
	}
	if e.SessionID != nil {
		t.Fatal("expected nil SessionID")
	}
}

// noopLog satisfies logger.Log for testing without importing zap.
type noopLog struct{}

func (noopLog) Debug(_ ...any)                                {}
func (noopLog) Info(_ ...any)                                 {}
func (noopLog) Warn(_ ...any)                                 {}
func (noopLog) Error(_ ...any)                                {}
func (noopLog) Fatal(_ ...any)                                {}
func (noopLog) Debugf(_ string, _ ...any)                     {}
func (noopLog) Infof(_ string, _ ...any)                      {}
func (noopLog) Warnf(_ string, _ ...any)                      {}
func (noopLog) Errorf(_ string, _ ...any)                     {}
func (noopLog) Fatalf(_ string, _ ...any)                     {}
func (noopLog) Debugw(_ string, _ ...any)                     {}
func (noopLog) Infow(_ string, _ ...any)                      {}
func (noopLog) Warnw(_ string, _ ...any)                      {}
func (noopLog) Errorw(_ string, _ ...any)                     {}
func (noopLog) Fatalw(_ string, _ ...any)                     {}
func (noopLog) Debugc(_ context.Context, _ string, _ ...any)  {}
func (noopLog) Infoc(_ context.Context, _ string, _ ...any)   {}
func (noopLog) Warnc(_ context.Context, _ string, _ ...any)   {}
func (noopLog) Errorc(_ context.Context, _ string, _ ...any)  {}
func (noopLog) Fatalc(_ context.Context, _ string, _ ...any)  {}
