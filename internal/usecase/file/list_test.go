package file_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFiles_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.FileMetadataFilter{
		Limit:  10,
		Offset: 0,
	}

	now := time.Now()
	mimeType := "image/jpeg"
	uploadedBy := "user-1"
	items := []domain.FileMetadata{
		{ID: "file-1", OriginalName: "photo.jpg", StoredName: "abc123.jpg", Bucket: "uploads", URL: "https://example.com/abc123.jpg", Size: 1024, MimeType: &mimeType, UploadedBy: &uploadedBy, CreatedAt: now, UpdatedAt: now},
		{ID: "file-2", OriginalName: "doc.pdf", StoredName: "def456.pdf", Bucket: "uploads", URL: "https://example.com/def456.pdf", Size: 2048, MimeType: &mimeType, UploadedBy: &uploadedBy, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("List", ctx, filter).Return(items, int64(2), nil)

	result, total, err := uc.ListFiles(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "file-1", result[0].ID)
	assert.Equal(t, "photo.jpg", result[0].OriginalName)
	repo.AssertExpectations(t)
}

func TestListFiles_Empty(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.FileMetadataFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.FileMetadata{}, int64(0), nil)

	result, total, err := uc.ListFiles(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestListFiles_WithSearchFilter(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.FileMetadataFilter{
		Search: "photo",
		Limit:  10,
	}

	repo.On("List", ctx, filter).Return([]domain.FileMetadata{}, int64(0), nil)

	_, _, err := uc.ListFiles(ctx, filter)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestListFiles_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.FileMetadataFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.FileMetadata{}, int64(0), errors.New("query failed"))

	_, _, err := uc.ListFiles(ctx, filter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	repo.AssertExpectations(t)
}
