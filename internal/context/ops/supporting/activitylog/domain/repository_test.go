package domain_test

import (
	"testing"

	"gct/internal/context/ops/supporting/activitylog/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestActivityLogFilter_ZeroValue(t *testing.T) {
	t.Parallel()

	filter := domain.ActivityLogFilter{}
	assert.Nil(t, filter.ActorID)
	assert.Nil(t, filter.EntityType)
	assert.Nil(t, filter.EntityID)
	assert.Nil(t, filter.FieldName)
	assert.Nil(t, filter.Action)
	assert.Nil(t, filter.FromDate)
	assert.Nil(t, filter.ToDate)
	assert.Nil(t, filter.Pagination)
}

func TestActivityLogView_Fields(t *testing.T) {
	t.Parallel()

	fn := "email"
	ov := "old@mail.com"
	nv := "new@mail.com"

	view := domain.ActivityLogView{
		ID:         1,
		ActorID:    uuid.New(),
		Action:     "user.updated",
		EntityType: "user",
		EntityID:   uuid.New(),
		FieldName:  &fn,
		OldValue:   &ov,
		NewValue:   &nv,
	}

	assert.Equal(t, int64(1), view.ID)
	assert.Equal(t, "user.updated", view.Action)
	assert.Equal(t, "email", *view.FieldName)
	assert.Equal(t, "old@mail.com", *view.OldValue)
	assert.Equal(t, "new@mail.com", *view.NewValue)
}
