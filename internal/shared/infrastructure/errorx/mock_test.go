package errorx

import (
	"context"
	"errors"
)

var ErrTestRepoFailure = errors.New("repository failure")

// mockRepository implements the Repository interface for testing
type mockRepository struct {
	err             error
	createCallCount int
	lastInput       LogErrorInput
}

func (m *mockRepository) Create(ctx context.Context, input LogErrorInput) error {
	m.createCallCount++
	m.lastInput = input
	return m.err
}

// mockLogger implements the logger.Log interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                      {}
func (m *mockLogger) Debugf(template string, args ...any)                    {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                {}
func (m *mockLogger) Info(args ...any)                                       {}
func (m *mockLogger) Infof(template string, args ...any)                     {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                 {}
func (m *mockLogger) Warn(args ...any)                                       {}
func (m *mockLogger) Warnf(template string, args ...any)                     {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                 {}
func (m *mockLogger) Error(args ...any)                                      {}
func (m *mockLogger) Errorf(template string, args ...any)                    {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                {}
func (m *mockLogger) Fatal(args ...any)                                      {}
func (m *mockLogger) Fatalf(template string, args ...any)                    {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, kv ...any)      {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, kv ...any)       {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, kv ...any)       {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, kv ...any)      {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, kv ...any)      {}
