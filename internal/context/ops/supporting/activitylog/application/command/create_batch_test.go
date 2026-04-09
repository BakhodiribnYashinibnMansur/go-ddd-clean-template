package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/ops/supporting/activitylog/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Repository ---

type mockWriteRepo struct {
	savedEntries []*domain.ActivityLogEntry
	saveErr      error
}

func (m *mockWriteRepo) SaveBatch(_ context.Context, entries []*domain.ActivityLogEntry) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedEntries = entries
	return nil
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

func TestCreateActivityLogBatchHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := NewCreateActivityLogBatchHandler(repo, &mockLogger{})

	entries := []*domain.ActivityLogEntry{
		domain.NewActivityLogEntry(uuid.New(), "user.updated", "user", uuid.New(), strPtr("email"), strPtr("old@m.com"), strPtr("new@m.com"), nil),
		domain.NewActivityLogEntry(uuid.New(), "user.updated", "user", uuid.New(), strPtr("username"), strPtr("old"), strPtr("new"), nil),
	}

	err := handler.Handle(context.Background(), CreateActivityLogBatchCommand{Entries: entries})
	require.NoError(t, err)
	assert.Len(t, repo.savedEntries, 2)
	assert.Equal(t, "email", *repo.savedEntries[0].FieldName())
	assert.Equal(t, "username", *repo.savedEntries[1].FieldName())
}

func TestCreateActivityLogBatchHandler_EmptyBatch(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := NewCreateActivityLogBatchHandler(repo, &mockLogger{})

	err := handler.Handle(context.Background(), CreateActivityLogBatchCommand{})
	require.NoError(t, err)
	assert.Nil(t, repo.savedEntries)
}

func TestCreateActivityLogBatchHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db write failed")
	repo := &mockWriteRepo{saveErr: repoErr}
	handler := NewCreateActivityLogBatchHandler(repo, &mockLogger{})

	entries := []*domain.ActivityLogEntry{
		domain.NewActivityLogEntry(uuid.New(), "user.created", "user", uuid.New(), strPtr("phone"), strPtr(""), strPtr("+998"), nil),
	}

	err := handler.Handle(context.Background(), CreateActivityLogBatchCommand{Entries: entries})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
}
