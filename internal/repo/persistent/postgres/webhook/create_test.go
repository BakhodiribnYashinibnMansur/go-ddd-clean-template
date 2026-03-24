package webhook

import (
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

func TestRepo_Create_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	w := &domain.Webhook{
		ID:       uuid.New(),
		Name:     "test-hook",
		URL:      "https://example.com/hook",
		Secret:   "secret123",
		Events:   []string{"order.created"},
		Headers:  map[string]any{"X-Key": "val"},
		IsActive: true,
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
		AddRow(now, now)

	mockPool.ExpectQuery("INSERT INTO webhooks").
		WithArgs(
			pgxmock.AnyArg(), // id
			pgxmock.AnyArg(), // name
			pgxmock.AnyArg(), // url
			pgxmock.AnyArg(), // secret
			pgxmock.AnyArg(), // events (json)
			pgxmock.AnyArg(), // headers (json)
			pgxmock.AnyArg(), // is_active
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, w)

	require.NoError(t, err)
	assert.Equal(t, now, w.CreatedAt)
	assert.Equal(t, now, w.UpdatedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Create_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	w := &domain.Webhook{
		ID:       uuid.New(),
		Name:     "fail-hook",
		URL:      "https://example.com/hook",
		Events:   []string{},
		Headers:  map[string]any{},
		IsActive: false,
	}

	mockPool.ExpectQuery("INSERT INTO webhooks").
		WithArgs(
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, w)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
