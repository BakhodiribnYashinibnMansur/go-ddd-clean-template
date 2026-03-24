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

func TestRepo_Update_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := "file-123"
	newName := "renamed.jpg"
	req := domain.UpdateFileMetadataRequest{
		OriginalName: &newName,
	}

	now := time.Now()
	mimeType := "image/jpeg"
	uploadedBy := "user-1"

	rows := pgxmock.NewRows([]string{
		"id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at",
	}).AddRow(id, newName, "abc.jpg", "uploads", "https://example.com/abc.jpg", int64(1024), &mimeType, &uploadedBy, now, now)

	mockPool.ExpectQuery("UPDATE " + table + " SET").
		WithArgs(newName, id). // original_name, NOW(), id
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Update(ctx, id, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newName, result.OriginalName)
	assert.Equal(t, id, result.ID)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := "file-456"
	newName := "renamed.jpg"
	req := domain.UpdateFileMetadataRequest{
		OriginalName: &newName,
	}

	mockPool.ExpectQuery("UPDATE " + table + " SET").
		WithArgs(newName, id).
		WillReturnError(errors.New("not found"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
