package emailtemplate

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Create_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	tmpl := &domain.EmailTemplate{
		ID:       "tmpl-uuid-1",
		Name:     "Welcome",
		Subject:  "Welcome!",
		HtmlBody: "<h1>Hi</h1>",
		TextBody: "Hi",
		Type:     "transactional",
		IsActive: true,
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
		AddRow(now, now)

	mockPool.ExpectQuery("INSERT INTO email_templates").
		WithArgs(
			tmpl.ID,
			tmpl.Name,
			tmpl.Subject,
			tmpl.HtmlBody,
			tmpl.TextBody,
			tmpl.Type,
			tmpl.IsActive,
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, tmpl)

	require.NoError(t, err)
	assert.Equal(t, now, tmpl.CreatedAt)
	assert.Equal(t, now, tmpl.UpdatedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Create_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	tmpl := &domain.EmailTemplate{
		ID:       "tmpl-uuid-2",
		Name:     "Fail",
		Subject:  "Subject",
		HtmlBody: "<p>Body</p>",
		TextBody: "Body",
		Type:     "marketing",
		IsActive: true,
	}

	mockPool.ExpectQuery("INSERT INTO email_templates").
		WithArgs(
			tmpl.ID,
			tmpl.Name,
			tmpl.Subject,
			tmpl.HtmlBody,
			tmpl.TextBody,
			tmpl.Type,
			tmpl.IsActive,
		).
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, tmpl)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
