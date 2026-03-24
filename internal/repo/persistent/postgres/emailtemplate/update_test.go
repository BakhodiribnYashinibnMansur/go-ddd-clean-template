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

func TestRepo_Update_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	tmpl := &domain.EmailTemplate{
		ID:       "tmpl-1",
		Name:     "Updated Name",
		Subject:  "Updated Subject",
		HtmlBody: "<p>Updated</p>",
		TextBody: "Updated",
		Type:     "marketing",
		IsActive: false,
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{"updated_at"}).AddRow(now)

	mockPool.ExpectQuery("UPDATE email_templates SET").
		WithArgs(
			tmpl.Name,
			tmpl.Subject,
			tmpl.HtmlBody,
			tmpl.TextBody,
			tmpl.Type,
			tmpl.IsActive,
			tmpl.ID,
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Update(ctx, tmpl)

	require.NoError(t, err)
	assert.Equal(t, now, tmpl.UpdatedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	tmpl := &domain.EmailTemplate{
		ID:       "tmpl-2",
		Name:     "Fail",
		Subject:  "Sub",
		HtmlBody: "<p>Body</p>",
		TextBody: "Body",
		Type:     "transactional",
		IsActive: true,
	}

	mockPool.ExpectQuery("UPDATE email_templates SET").
		WithArgs(
			tmpl.Name,
			tmpl.Subject,
			tmpl.HtmlBody,
			tmpl.TextBody,
			tmpl.Type,
			tmpl.IsActive,
			tmpl.ID,
		).
		WillReturnError(errors.New("update failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Update(ctx, tmpl)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
