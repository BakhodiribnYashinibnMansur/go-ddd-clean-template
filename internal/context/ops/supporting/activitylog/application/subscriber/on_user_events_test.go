package subscriber

import (
	"context"
	"testing"

	contractevents "gct/internal/contract/events"
	"gct/internal/context/ops/supporting/activitylog/application/command"
	"gct/internal/context/ops/supporting/activitylog/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Write Repo ---

type mockWriteRepo struct {
	savedEntries []*domain.ActivityLogEntry
}

func (m *mockWriteRepo) SaveBatch(_ context.Context, entries []*domain.ActivityLogEntry) error {
	m.savedEntries = append(m.savedEntries, entries...)
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

func TestHandleUserProfileUpdatedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	actorID := uuid.New()
	userID := uuid.New()

	event := contractevents.NewUserProfileUpdatedV2(userID, actorID, []contractevents.FieldChange{
		{FieldName: "email", OldValue: "old@mail.com", NewValue: "new@mail.com"},
		{FieldName: "username", OldValue: "john", NewValue: "john_doe"},
	})

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 2)

	assert.Equal(t, "user.updated", repo.savedEntries[0].Action())
	assert.Equal(t, "user", repo.savedEntries[0].EntityType())
	assert.Equal(t, userID, repo.savedEntries[0].EntityID())
	assert.Equal(t, actorID, repo.savedEntries[0].ActorID())
	assert.Equal(t, "email", *repo.savedEntries[0].FieldName())
	assert.Equal(t, "old@mail.com", *repo.savedEntries[0].OldValue())
	assert.Equal(t, "new@mail.com", *repo.savedEntries[0].NewValue())

	assert.Equal(t, "username", *repo.savedEntries[1].FieldName())
	assert.Equal(t, "john", *repo.savedEntries[1].OldValue())
	assert.Equal(t, "john_doe", *repo.savedEntries[1].NewValue())
}

func TestHandleUserCreatedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	actorID := uuid.New()
	userID := uuid.New()

	event := contractevents.NewUserCreatedV2(userID, actorID, []contractevents.FieldChange{
		{FieldName: "phone", OldValue: "", NewValue: "+998901234567"},
		{FieldName: "password", OldValue: contractevents.RedactedValue, NewValue: contractevents.RedactedValue},
	})

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 2)
	assert.Equal(t, "user.created", repo.savedEntries[0].Action())
	assert.Equal(t, "phone", *repo.savedEntries[0].FieldName())
	assert.Equal(t, "", *repo.savedEntries[0].OldValue())
	assert.Equal(t, "+998901234567", *repo.savedEntries[0].NewValue())

	assert.Equal(t, "password", *repo.savedEntries[1].FieldName())
	assert.Equal(t, contractevents.RedactedValue, *repo.savedEntries[1].OldValue())
}

func TestHandleUserDeletedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	actorID := uuid.New()
	userID := uuid.New()

	event := contractevents.NewUserDeletedV2(userID, actorID)

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 1)
	assert.Equal(t, "user.deleted", repo.savedEntries[0].Action())
	assert.Equal(t, actorID, repo.savedEntries[0].ActorID())
	assert.Equal(t, userID, repo.savedEntries[0].EntityID())
	assert.Nil(t, repo.savedEntries[0].FieldName())
}

func TestHandleUserRoleChangedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	actorID := uuid.New()
	userID := uuid.New()
	oldRole := uuid.New()
	newRole := uuid.New()

	event := contractevents.NewUserRoleChangedV2(userID, actorID, []contractevents.FieldChange{
		{FieldName: "role_id", OldValue: oldRole.String(), NewValue: newRole.String()},
	})

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 1)
	assert.Equal(t, "user.role_changed", repo.savedEntries[0].Action())
	assert.Equal(t, "role_id", *repo.savedEntries[0].FieldName())
	assert.Equal(t, oldRole.String(), *repo.savedEntries[0].OldValue())
	assert.Equal(t, newRole.String(), *repo.savedEntries[0].NewValue())
}

func TestHandleUserApprovedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	event := contractevents.NewUserApprovedV2(uuid.New(), uuid.New(), []contractevents.FieldChange{
		{FieldName: "is_approved", OldValue: "false", NewValue: "true"},
	})

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 1)
	assert.Equal(t, "user.approved", repo.savedEntries[0].Action())
	assert.Equal(t, "is_approved", *repo.savedEntries[0].FieldName())
}

func TestHandleUserPasswordChangedV2(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	event := contractevents.NewUserPasswordChangedV2(uuid.New(), uuid.New())

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, repo.savedEntries, 1)
	assert.Equal(t, "user.password_changed", repo.savedEntries[0].Action())
	assert.Equal(t, "password", *repo.savedEntries[0].FieldName())
	assert.Equal(t, contractevents.RedactedValue, *repo.savedEntries[0].OldValue())
	assert.Equal(t, contractevents.RedactedValue, *repo.savedEntries[0].NewValue())
}

func TestHandleUnknownEvent_NoEntries(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewUserEventSubscriber(handler, &mockLogger{})

	// An event type that the subscriber doesn't handle should produce no entries.
	event := contractevents.UserDeactivatedV1{
		BaseEvent: contractevents.BaseEvent{
			Envelope: contractevents.NewEnvelope("user.deactivated.v1", uuid.New(), 1),
		},
	}

	err := sub.handle(context.Background(), event)
	require.NoError(t, err)
	assert.Empty(t, repo.savedEntries)
}
