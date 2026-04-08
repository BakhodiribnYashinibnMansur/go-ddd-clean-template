package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/supporting/activitylog/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Read Repo ---

type mockReadRepo struct {
	views []*domain.ActivityLogView
	total int64
	err   error
}

func (m *mockReadRepo) List(_ context.Context, _ domain.ActivityLogFilter) ([]*domain.ActivityLogView, int64, error) {
	return m.views, m.total, m.err
}

// --- Mock Logger ---

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                            {}
func (m *mockLogger) Debugf(template string, args ...any)                          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Info(args ...any)                                             {}
func (m *mockLogger) Infof(template string, args ...any)                           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Warn(args ...any)                                             {}
func (m *mockLogger) Warnf(template string, args ...any)                           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Error(args ...any)                                            {}
func (m *mockLogger) Errorf(template string, args ...any)                          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Fatal(args ...any)                                            {}
func (m *mockLogger) Fatalf(template string, args ...any)                          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, keysAndValues ...any) {}

// --- Tests ---

func strPtr(s string) *string { return &s }

func TestListActivityLogsHandler_Success(t *testing.T) {
	t.Parallel()

	entityID := uuid.New()
	views := []*domain.ActivityLogView{
		{
			ID:         1,
			ActorID:    uuid.New(),
			Action:     "user.updated",
			EntityType: "user",
			EntityID:   entityID,
			FieldName:  strPtr("email"),
			OldValue:   strPtr("old@m.com"),
			NewValue:   strPtr("new@m.com"),
			CreatedAt:  time.Now(),
		},
		{
			ID:         2,
			ActorID:    uuid.New(),
			Action:     "user.updated",
			EntityType: "user",
			EntityID:   entityID,
			FieldName:  strPtr("username"),
			OldValue:   strPtr("john"),
			NewValue:   strPtr("john_doe"),
			CreatedAt:  time.Now(),
		},
	}

	repo := &mockReadRepo{views: views, total: 2}
	handler := NewListActivityLogsHandler(repo, &mockLogger{})

	result, err := handler.Handle(context.Background(), ListActivityLogsQuery{
		Filter: domain.ActivityLogFilter{
			EntityID: &entityID,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, "email", *result.Entries[0].FieldName)
	assert.Equal(t, "username", *result.Entries[1].FieldName)
}

func TestListActivityLogsHandler_Empty(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{views: nil, total: 0}
	handler := NewListActivityLogsHandler(repo, &mockLogger{})

	result, err := handler.Handle(context.Background(), ListActivityLogsQuery{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
	assert.Empty(t, result.Entries)
}

func TestListActivityLogsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{err: errors.New("query failed")}
	handler := NewListActivityLogsHandler(repo, &mockLogger{})

	result, err := handler.Handle(context.Background(), ListActivityLogsQuery{})
	assert.Error(t, err)
	assert.Nil(t, result)
}
