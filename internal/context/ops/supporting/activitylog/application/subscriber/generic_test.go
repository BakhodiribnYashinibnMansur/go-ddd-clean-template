package subscriber

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/ops/supporting/activitylog/application/command"
	"gct/internal/context/ops/supporting/activitylog/domain"
	shareddomain "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake event types for testing ---

// plainEvent has no metadata.
type plainEvent struct {
	id   uuid.UUID
	name string
}

func (e plainEvent) EventName() string      { return e.name }
func (e plainEvent) OccurredAt() time.Time  { return time.Now() }
func (e plainEvent) AggregateID() uuid.UUID { return e.id }

// metadataEvent implements MetadataProvider.
type metadataEvent struct {
	plainEvent
	meta string
}

func (e metadataEvent) ActivityMetadata() string { return e.meta }

// Compile-time check.
var _ shareddomain.MetadataProvider = metadataEvent{}

// --- Tests ---

func TestGenericSubscriber_PlainEvent(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewGenericSubscriber(handler, &mockLogger{})

	userID := uuid.New()
	mapping := eventMapping{Event: "user.deactivated", Action: "user.deactivated", EntityType: "user"}
	event := plainEvent{id: userID, name: "user.deactivated"}

	err := sub.handle(context.Background(), mapping, event)
	require.NoError(t, err)

	require.Len(t, repo.savedEntries, 1)
	entry := repo.savedEntries[0]
	assert.Equal(t, "user.deactivated", entry.Action())
	assert.Equal(t, "user", entry.EntityType())
	assert.Equal(t, userID, entry.EntityID())
	assert.Equal(t, userID, entry.ActorID()) // actor = aggregate for V1
	assert.Nil(t, entry.FieldName())
	assert.Nil(t, entry.OldValue())
	assert.Nil(t, entry.NewValue())
	assert.Nil(t, entry.Metadata()) // no metadata
}

func TestGenericSubscriber_EventWithMetadata(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewGenericSubscriber(handler, &mockLogger{})

	userID := uuid.New()
	sessionID := uuid.New()
	mapping := eventMapping{Event: "user.signed_in", Action: "user.signed_in", EntityType: "user"}
	event := metadataEvent{
		plainEvent: plainEvent{id: userID, name: "user.signed_in"},
		meta:       "ip=192.168.1.1 session=" + sessionID.String(),
	}

	err := sub.handle(context.Background(), mapping, event)
	require.NoError(t, err)

	require.Len(t, repo.savedEntries, 1)
	entry := repo.savedEntries[0]
	assert.Equal(t, "user.signed_in", entry.Action())
	assert.Equal(t, "user", entry.EntityType())
	require.NotNil(t, entry.Metadata())
	assert.Contains(t, *entry.Metadata(), "ip=192.168.1.1")
	assert.Contains(t, *entry.Metadata(), "session="+sessionID.String())
}

func TestGenericSubscriber_EmptyMetadataIsNil(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewGenericSubscriber(handler, &mockLogger{})

	mapping := eventMapping{Event: "test.event", Action: "test.action", EntityType: "test"}
	event := metadataEvent{
		plainEvent: plainEvent{id: uuid.New(), name: "test.event"},
		meta:       "", // empty metadata should be treated as nil
	}

	err := sub.handle(context.Background(), mapping, event)
	require.NoError(t, err)

	require.Len(t, repo.savedEntries, 1)
	assert.Nil(t, repo.savedEntries[0].Metadata())
}

func TestGenericSubscriber_RepoErrorDoesNotPropagate(t *testing.T) {
	t.Parallel()

	// The generic subscriber should swallow repo errors (best-effort logging).
	repo := &mockWriteRepoWithError{}
	handler := command.NewCreateActivityLogBatchHandler(repo, &mockLogger{})
	sub := NewGenericSubscriber(handler, &mockLogger{})

	mapping := eventMapping{Event: "role.created", Action: "role.created", EntityType: "role"}
	event := plainEvent{id: uuid.New(), name: "role.created"}

	err := sub.handle(context.Background(), mapping, event)
	assert.NoError(t, err) // error swallowed
}

type mockWriteRepoWithError struct{}

func (m *mockWriteRepoWithError) SaveBatch(_ context.Context, _ []*domain.ActivityLogEntry) error {
	return assert.AnError
}

func TestGenericSubscriber_MappingCount(t *testing.T) {
	t.Parallel()
	// Ensure we have mappings for all expected events.
	assert.GreaterOrEqual(t, len(eventMappings), 30, "expected at least 30 event mappings")
}
