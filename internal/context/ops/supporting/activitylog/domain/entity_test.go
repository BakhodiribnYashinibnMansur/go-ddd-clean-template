package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/ops/supporting/activitylog/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func TestNewActivityLogEntry_AllFields(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	entityID := uuid.New()
	fieldName := "email"
	oldValue := "old@mail.com"
	newValue := "new@mail.com"
	meta := "some context"

	entry := domain.NewActivityLogEntry(
		actorID, "user.updated", "user", entityID,
		&fieldName, &oldValue, &newValue, &meta,
	)

	assert.Equal(t, actorID, entry.ActorID())
	assert.Equal(t, "user.updated", entry.Action())
	assert.Equal(t, "user", entry.EntityType())
	assert.Equal(t, entityID, entry.EntityID())
	assert.Equal(t, &fieldName, entry.FieldName())
	assert.Equal(t, &oldValue, entry.OldValue())
	assert.Equal(t, &newValue, entry.NewValue())
	assert.Equal(t, &meta, entry.Metadata())
	assert.False(t, entry.CreatedAt().IsZero())
}

func TestNewActivityLogEntry_NilOptionalFields(t *testing.T) {
	t.Parallel()

	entry := domain.NewActivityLogEntry(
		uuid.New(), "user.deleted", "user", uuid.New(),
		nil, nil, nil, nil,
	)

	assert.Nil(t, entry.FieldName())
	assert.Nil(t, entry.OldValue())
	assert.Nil(t, entry.NewValue())
	assert.Nil(t, entry.Metadata())
}

func TestReconstructActivityLogEntry(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	entityID := uuid.New()
	now := time.Now().UTC()
	fn := "username"

	entry := domain.ReconstructActivityLogEntry(
		42, actorID, "user.updated", "user", entityID,
		&fn, strPtr("old"), strPtr("new"), nil, now,
	)

	assert.Equal(t, int64(42), entry.ID())
	assert.Equal(t, actorID, entry.ActorID())
	assert.Equal(t, "user.updated", entry.Action())
	assert.Equal(t, entityID, entry.EntityID())
	assert.Equal(t, &fn, entry.FieldName())
	assert.Equal(t, now, entry.CreatedAt())
}

func TestNewActivityLogEntry_CreateAction(t *testing.T) {
	t.Parallel()

	fn := "phone"
	nv := "+998901234567"

	entry := domain.NewActivityLogEntry(
		uuid.New(), "user.created", "user", uuid.New(),
		&fn, strPtr(""), &nv, nil,
	)

	assert.Equal(t, "user.created", entry.Action())
	assert.Equal(t, "", *entry.OldValue())
	assert.Equal(t, "+998901234567", *entry.NewValue())
}
