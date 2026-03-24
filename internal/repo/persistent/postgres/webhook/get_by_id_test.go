package webhook

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_GetByID_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()
	now := time.Now()
	events := []string{"order.created", "order.updated"}
	headers := map[string]any{"X-Custom": "value"}
	eventsJSON, _ := json.Marshal(events)
	headersJSON, _ := json.Marshal(headers)

	rows := pgxmock.NewRows([]string{
		"id", "name", "url", "secret", "events", "headers",
		"is_active", "created_at", "updated_at", "deleted_at",
	}).AddRow(id, "test-hook", "https://example.com/hook", "secret",
		eventsJSON, headersJSON, true, now, now, nil)

	// squirrel converts uuid.UUID to string in args
	mockPool.ExpectQuery("SELECT (.+) FROM webhooks").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "test-hook", result.Name)
	assert.Equal(t, "https://example.com/hook", result.URL)
	assert.Equal(t, "secret", result.Secret)
	assert.Equal(t, events, result.Events)
	assert.True(t, result.IsActive)
	assert.Equal(t, now, result.CreatedAt)
	assert.Nil(t, result.DeletedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_GetByID_NotFound(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()

	mockPool.ExpectQuery("SELECT (.+) FROM webhooks").
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(errors.New("no rows in result set"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByID(ctx, id)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
