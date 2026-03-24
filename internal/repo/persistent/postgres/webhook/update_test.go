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

func TestRepo_Update_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	now := time.Now()
	w := &domain.Webhook{
		ID:       uuid.New(),
		Name:     "updated-hook",
		URL:      "https://example.com/updated",
		Secret:   "new-secret",
		Events:   []string{"order.created", "order.updated"},
		Headers:  map[string]any{"X-New": "header"},
		IsActive: true,
	}

	rows := pgxmock.NewRows([]string{"updated_at"}).AddRow(now)

	// SET name=$1, url=$2, secret=$3, events=$4, headers=$5, is_active=$6, updated_at=NOW() WHERE id=$7
	// NOW() is Expr => 7 args total
	mockPool.ExpectQuery("UPDATE webhooks SET").
		WithArgs(
			pgxmock.AnyArg(), // name
			pgxmock.AnyArg(), // url
			pgxmock.AnyArg(), // secret
			pgxmock.AnyArg(), // events (json)
			pgxmock.AnyArg(), // headers (json)
			pgxmock.AnyArg(), // is_active
			pgxmock.AnyArg(), // id (WHERE)
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Update(ctx, w)

	require.NoError(t, err)
	assert.Equal(t, now, w.UpdatedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	w := &domain.Webhook{
		ID:       uuid.New(),
		Name:     "fail-update",
		URL:      "https://example.com/hook",
		Events:   []string{},
		Headers:  map[string]any{},
		IsActive: false,
	}

	mockPool.ExpectQuery("UPDATE webhooks SET").
		WithArgs(
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnError(errors.New("update failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Update(ctx, w)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
