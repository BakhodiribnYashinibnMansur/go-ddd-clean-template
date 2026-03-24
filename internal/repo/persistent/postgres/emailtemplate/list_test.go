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

func TestRepo_List_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.EmailTemplateFilter{
		Limit:  10,
		Offset: 0,
	}

	// Count subquery
	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// Main query
	now := time.Now()
	listRows := pgxmock.NewRows([]string{
		"id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at",
	}).
		AddRow("id-1", "Template 1", "Sub 1", "<p>1</p>", "1", "transactional", true, now, now).
		AddRow("id-2", "Template 2", "Sub 2", "<p>2</p>", "2", "marketing", false, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
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
	assert.Equal(t, "Template 1", items[0].Name)
	assert.Equal(t, "Template 2", items[1].Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.EmailTemplateFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at",
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
	assert.NotNil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_CountError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.EmailTemplateFilter{Limit: 10}

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

	filter := domain.EmailTemplateFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
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

func TestRepo_List_WithSearchFilter(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.EmailTemplateFilter{
		Search: "welcome",
		Limit:  10,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WithArgs("%welcome%").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

	now := time.Now()
	mockPool.ExpectQuery("SELECT (.+) FROM email_templates").
		WithArgs("%welcome%").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at",
		}).AddRow("id-1", "Welcome Email", "Welcome!", "<h1>Hi</h1>", "Hi", "transactional", true, now, now))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Welcome Email", items[0].Name)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
