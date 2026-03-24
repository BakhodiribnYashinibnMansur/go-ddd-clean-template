package emailtemplate

import (
	"errors"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_GetByID_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at",
	}).AddRow("tmpl-1", "Welcome", "Welcome!", "<h1>Hi</h1>", "Hi", "transactional", true, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
		WithArgs("tmpl-1").
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByID(ctx, "tmpl-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tmpl-1", result.ID)
	assert.Equal(t, "Welcome", result.Name)
	assert.Equal(t, "Welcome!", result.Subject)
	assert.Equal(t, "<h1>Hi</h1>", result.HtmlBody)
	assert.Equal(t, "Hi", result.TextBody)
	assert.Equal(t, "transactional", result.Type)
	assert.True(t, result.IsActive)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_GetByID_NotFound(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
		WithArgs("nonexistent").
		WillReturnError(errors.New("no rows in result set"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByID(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
