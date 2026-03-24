package webhook

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_List_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.WebhookFilter{
		Limit:  10,
		Offset: 0,
	}

	// deleted_at IS NULL => 0 args for count subquery
	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// Main query: deleted_at IS NULL, LIMIT/OFFSET inlined by squirrel => 0 args
	id1, id2 := uuid.New(), uuid.New()
	now := time.Now()
	events1, _ := json.Marshal([]string{"order.created"})
	headers1, _ := json.Marshal(map[string]any{"X-Key": "val"})
	events2, _ := json.Marshal([]string{})
	headers2, _ := json.Marshal(map[string]any{})

	listRows := pgxmock.NewRows([]string{
		"id", "name", "url", "secret", "events", "headers",
		"is_active", "created_at", "updated_at",
	}).
		AddRow(id1, "hook-1", "https://example.com/1", "s1", events1, headers1, true, now, now).
		AddRow(id2, "hook-2", "https://example.com/2", "", events2, headers2, false, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM webhooks").
		WillReturnRows(listRows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, "hook-1", items[0].Name)
	assert.Equal(t, "hook-2", items[1].Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.WebhookFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM webhooks").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "url", "secret", "events", "headers",
			"is_active", "created_at", "updated_at",
		}))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	assert.NotNil(t, items) // should be [] not nil
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_CountError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.WebhookFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count query failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_QueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.WebhookFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mockPool.ExpectQuery("SELECT (.+) FROM webhooks").
		WillReturnError(errors.New("query failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
