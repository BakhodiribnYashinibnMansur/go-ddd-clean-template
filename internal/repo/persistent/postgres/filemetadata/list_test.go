package filemetadata

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

	filter := domain.FileMetadataFilter{
		Limit:  10,
		Offset: 0,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	now := time.Now()
	mimeType := "image/jpeg"
	uploadedBy := "user-1"

	listRows := pgxmock.NewRows([]string{
		"id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at",
	}).
		AddRow("f-1", "photo.jpg", "abc.jpg", "uploads", "https://example.com/abc.jpg", int64(1024), &mimeType, &uploadedBy, now, now).
		AddRow("f-2", "doc.pdf", "def.pdf", "uploads", "https://example.com/def.pdf", int64(2048), &mimeType, &uploadedBy, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
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
	assert.Equal(t, "photo.jpg", items[0].OriginalName)
	assert.Equal(t, "doc.pdf", items[1].OriginalName)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.FileMetadataFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at",
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

func TestRepo_List_WithSearchFilter(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.FileMetadataFilter{
		Search: "photo",
		Limit:  10,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WithArgs("%photo%").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at",
		}))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_WithMimeTypeFilter(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.FileMetadataFilter{
		MimeType: "image/jpeg",
		Limit:    10,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WithArgs("image/jpeg").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at",
		}))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_CountError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.FileMetadataFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count failed"))

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

	filter := domain.FileMetadataFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
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
