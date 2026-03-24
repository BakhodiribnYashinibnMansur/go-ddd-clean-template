package iprule_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.IPRuleFilter{
		Search: "192.168",
		Limit:  10,
		Offset: 0,
	}

	expected := []domain.IPRule{
		{
			ID:        uuid.New(),
			IPAddress: "192.168.1.1",
			Type:      "whitelist",
			Reason:    "Office",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			IPAddress: "192.168.1.2",
			Type:      "whitelist",
			Reason:    "VPN",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	repo.On("List", ctx, filter).Return(expected, int64(2), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, expected[0].IPAddress, items[0].IPAddress)
	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.IPRuleFilter{
		Search: "nonexistent",
		Limit:  10,
	}

	repo.On("List", ctx, filter).Return([]domain.IPRule{}, int64(0), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	repo.AssertExpectations(t)
}

func TestList_Error(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.IPRuleFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.IPRule{}, int64(0), errors.New("database error"))

	items, total, err := uc.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
