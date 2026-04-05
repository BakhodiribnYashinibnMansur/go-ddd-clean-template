package asynq

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// mockLogger implements logger.Log for testing purposes.
type mockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
	fatalMsgs []string
}

func (m *mockLogger) Debug(args ...any)                                    { m.debugMsgs = append(m.debugMsgs, fmt.Sprint(args...)) }
func (m *mockLogger) Debugf(template string, args ...any)                  { m.debugMsgs = append(m.debugMsgs, fmt.Sprintf(template, args...)) }
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)              { m.debugMsgs = append(m.debugMsgs, msg) }
func (m *mockLogger) Info(args ...any)                                     { m.infoMsgs = append(m.infoMsgs, fmt.Sprint(args...)) }
func (m *mockLogger) Infof(template string, args ...any)                   { m.infoMsgs = append(m.infoMsgs, fmt.Sprintf(template, args...)) }
func (m *mockLogger) Infow(msg string, keysAndValues ...any)               { m.infoMsgs = append(m.infoMsgs, msg) }
func (m *mockLogger) Warn(args ...any)                                     { m.warnMsgs = append(m.warnMsgs, fmt.Sprint(args...)) }
func (m *mockLogger) Warnf(template string, args ...any)                   { m.warnMsgs = append(m.warnMsgs, fmt.Sprintf(template, args...)) }
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)               { m.warnMsgs = append(m.warnMsgs, msg) }
func (m *mockLogger) Error(args ...any)                                    { m.errorMsgs = append(m.errorMsgs, fmt.Sprint(args...)) }
func (m *mockLogger) Errorf(template string, args ...any)                  { m.errorMsgs = append(m.errorMsgs, fmt.Sprintf(template, args...)) }
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)              { m.errorMsgs = append(m.errorMsgs, msg) }
func (m *mockLogger) Fatal(args ...any)                                    { m.fatalMsgs = append(m.fatalMsgs, fmt.Sprint(args...)) }
func (m *mockLogger) Fatalf(template string, args ...any)                  { m.fatalMsgs = append(m.fatalMsgs, fmt.Sprintf(template, args...)) }
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)              { m.fatalMsgs = append(m.fatalMsgs, msg) }
func (m *mockLogger) Debugc(_ context.Context, msg string, _ ...any)       { m.debugMsgs = append(m.debugMsgs, msg) }
func (m *mockLogger) Infoc(_ context.Context, msg string, _ ...any)        { m.infoMsgs = append(m.infoMsgs, msg) }
func (m *mockLogger) Warnc(_ context.Context, msg string, _ ...any)        { m.warnMsgs = append(m.warnMsgs, msg) }
func (m *mockLogger) Errorc(_ context.Context, msg string, _ ...any)       { m.errorMsgs = append(m.errorMsgs, msg) }
func (m *mockLogger) Fatalc(_ context.Context, msg string, _ ...any)       { m.fatalMsgs = append(m.fatalMsgs, msg) }

// ---------------------------------------------------------------------------
// AsynqLogger tests
// ---------------------------------------------------------------------------

func TestNewAsynqLogger(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)
	if al == nil {
		t.Fatal("NewAsynqLogger returned nil")
	}
	if al.log != ml {
		t.Fatal("NewAsynqLogger did not store the provided logger")
	}
}

func TestAsynqLogger_Debug(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Debug("test debug message")

	if len(ml.debugMsgs) != 1 {
		t.Fatalf("expected 1 debug message, got %d", len(ml.debugMsgs))
	}
	if ml.debugMsgs[0] != "test debug message" {
		t.Fatalf("expected 'test debug message', got %q", ml.debugMsgs[0])
	}
}

func TestAsynqLogger_DebugMultipleArgs(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Debug("hello", " ", "world")

	if len(ml.debugMsgs) != 1 {
		t.Fatalf("expected 1 debug message, got %d", len(ml.debugMsgs))
	}
	// fmt.Sprint joins without separator
	if ml.debugMsgs[0] != "hello world" {
		t.Fatalf("expected 'hello world', got %q", ml.debugMsgs[0])
	}
}

func TestAsynqLogger_Info(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Info("server started")

	if len(ml.infoMsgs) != 1 {
		t.Fatalf("expected 1 info message, got %d", len(ml.infoMsgs))
	}
	want := "\u2139\ufe0f  Asynq: server started"
	if ml.infoMsgs[0] != want {
		t.Fatalf("expected %q, got %q", want, ml.infoMsgs[0])
	}
}

func TestAsynqLogger_InfoPrefix(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Info("processing tasks")

	if !strings.HasPrefix(ml.infoMsgs[0], "\u2139\ufe0f  Asynq: ") {
		t.Fatalf("info message should have prefix, got %q", ml.infoMsgs[0])
	}
}

func TestAsynqLogger_InfoEmptyMessage(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	// Info with empty string should not log (see source: len(msg) > 0 check)
	al.Info("")

	if len(ml.infoMsgs) != 0 {
		t.Fatalf("expected 0 info messages for empty input, got %d", len(ml.infoMsgs))
	}
}

func TestAsynqLogger_Warn(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Warn("slow task")

	if len(ml.warnMsgs) != 1 {
		t.Fatalf("expected 1 warn message, got %d", len(ml.warnMsgs))
	}
	want := "\u26a0\ufe0f  Asynq warning: slow task"
	if ml.warnMsgs[0] != want {
		t.Fatalf("expected %q, got %q", want, ml.warnMsgs[0])
	}
}

func TestAsynqLogger_WarnPrefix(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Warn("something")

	if !strings.HasPrefix(ml.warnMsgs[0], "\u26a0\ufe0f  Asynq warning: ") {
		t.Fatalf("warn message should have prefix, got %q", ml.warnMsgs[0])
	}
}

func TestAsynqLogger_Error(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Error("connection lost")

	if len(ml.errorMsgs) != 1 {
		t.Fatalf("expected 1 error message, got %d", len(ml.errorMsgs))
	}
	want := "\u274c Asynq error: connection lost"
	if ml.errorMsgs[0] != want {
		t.Fatalf("expected %q, got %q", want, ml.errorMsgs[0])
	}
}

func TestAsynqLogger_ErrorPrefix(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Error("something broke")

	if !strings.HasPrefix(ml.errorMsgs[0], "\u274c Asynq error: ") {
		t.Fatalf("error message should have prefix, got %q", ml.errorMsgs[0])
	}
}

func TestAsynqLogger_Fatal(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Fatal("critical failure")

	if len(ml.fatalMsgs) != 1 {
		t.Fatalf("expected 1 fatal message, got %d", len(ml.fatalMsgs))
	}
	want := "\U0001f480 Asynq fatal: critical failure"
	if ml.fatalMsgs[0] != want {
		t.Fatalf("expected %q, got %q", want, ml.fatalMsgs[0])
	}
}

func TestAsynqLogger_FatalPrefix(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	al.Fatal("boom")

	if !strings.HasPrefix(ml.fatalMsgs[0], "\U0001f480 Asynq fatal: ") {
		t.Fatalf("fatal message should have prefix, got %q", ml.fatalMsgs[0])
	}
}

func TestAsynqLogger_EmptyMessageHandledGracefully(t *testing.T) {
	ml := &mockLogger{}
	al := NewAsynqLogger(ml)

	// None of these should panic
	al.Debug()
	al.Info()
	al.Warn()
	al.Error()
	al.Fatal()

	// Debug gets empty string from fmt.Sprint()
	if len(ml.debugMsgs) != 1 {
		t.Fatalf("expected 1 debug message, got %d", len(ml.debugMsgs))
	}
	// Info checks len(msg) > 0, so empty fmt.Sprint() should not log
	if len(ml.infoMsgs) != 0 {
		t.Fatalf("expected 0 info messages for no-args, got %d", len(ml.infoMsgs))
	}
	// Warn, Error, Fatal still log with prefix + empty message
	if len(ml.warnMsgs) != 1 {
		t.Fatalf("expected 1 warn message, got %d", len(ml.warnMsgs))
	}
	if len(ml.errorMsgs) != 1 {
		t.Fatalf("expected 1 error message, got %d", len(ml.errorMsgs))
	}
	if len(ml.fatalMsgs) != 1 {
		t.Fatalf("expected 1 fatal message, got %d", len(ml.fatalMsgs))
	}
}

// ---------------------------------------------------------------------------
// RegisterExternalHandlers tests
// ---------------------------------------------------------------------------

func TestRegisterExternalHandlers_NilFCM(t *testing.T) {
	ml := &mockLogger{}
	w := &Worker{
		mux: nil, // won't be touched since fcm is nil
		log: ml,
	}

	// Need a real mux for the telegram path in case tg is not nil,
	// but here both are nil so mux is never accessed.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil FCM sender caused panic: %v", r)
		}
	}()

	w.RegisterExternalHandlers(nil, nil)
}

func TestRegisterExternalHandlers_NilTelegram(t *testing.T) {
	ml := &mockLogger{}
	w := &Worker{
		mux: nil,
		log: ml,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil Telegram sender caused panic: %v", r)
		}
	}()

	w.RegisterExternalHandlers(nil, nil)
}

func TestRegisterExternalHandlers_BothNil(t *testing.T) {
	ml := &mockLogger{}
	w := &Worker{
		mux: nil,
		log: ml,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("both nil senders caused panic: %v", r)
		}
	}()

	w.RegisterExternalHandlers(nil, nil)

	// No handlers should have been registered, so no log messages about registration
	for _, msg := range ml.infoMsgs {
		if strings.Contains(msg, "Registered") {
			t.Fatalf("unexpected registration message when both senders are nil: %q", msg)
		}
	}
}
