package errorx

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestLogError_AppError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	appErr := New(ErrInternal, "something failed")
	appErr.Details = "detailed info"
	appErr.WithField("key1", "value1")

	// Should not panic
	LogError(logger, appErr)
}

func TestLogError_StandardError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	err := errors.New("standard error")
	// Should not panic
	LogError(logger, err)
}

func TestLogError_AppErrorWithWrappedError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	origErr := errors.New("original")
	appErr := Wrap(origErr, ErrDatabase, "database failed")

	// Should not panic
	LogError(logger, appErr)
}

func TestLogWarn_AppError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	appErr := New(ErrBadRequest, "bad input")
	appErr.WithField("field", "email")

	// Should not panic
	LogWarn(logger, appErr)
}

func TestLogWarn_StandardError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	err := errors.New("standard warning")
	// Should not panic
	LogWarn(logger, err)
}

func TestLogInfo_AppError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	appErr := New(ErrNotFound, "resource missing")

	// Should not panic
	LogInfo(logger, appErr, "info about missing resource")
}

func TestLogInfo_StandardError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	err := errors.New("standard info")
	// Should not panic
	LogInfo(logger, err, "some info")
}

func TestSetReporter(t *testing.T) {
	// Save original reporter and restore after test
	origReporter := getReporter()
	defer func() {
		if origReporter != nil {
			SetReporter(origReporter)
		} else {
			reporterPtr.Store(nil)
		}
	}()

	called := false
	mockReporter := &mockReporterImpl{
		sendFunc: func(err error) error {
			called = true
			return nil
		},
	}

	SetReporter(mockReporter)

	logger := zaptest.NewLogger(t)
	LogError(logger, errors.New("test error"))

	if !called {
		t.Error("expected reporter SendError to be called")
	}
}

func TestLogError_WithNilReporter(t *testing.T) {
	origReporter := getReporter()
	defer func() {
		if origReporter != nil {
			SetReporter(origReporter)
		} else {
			reporterPtr.Store(nil)
		}
	}()
	reporterPtr.Store(nil)

	logger := zaptest.NewLogger(t)
	// Should not panic when reporter is nil
	LogError(logger, errors.New("test error"))
}

func TestLogError_WithNoopLogger(t *testing.T) {
	logger := zap.NewNop()
	// Should not panic with noop logger
	LogError(logger, New(ErrInternal, "test"))
	LogWarn(logger, New(ErrBadRequest, "test"))
	LogInfo(logger, New(ErrNotFound, "test"), "info")
}

type mockReporterImpl struct {
	sendFunc func(err error) error
}

func (m *mockReporterImpl) SendError(err error) error {
	return m.sendFunc(err)
}
