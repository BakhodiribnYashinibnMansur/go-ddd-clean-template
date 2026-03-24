package file_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFile_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := "file-123"
	newName := "renamed.jpg"
	req := domain.UpdateFileMetadataRequest{
		OriginalName: &newName,
	}

	now := time.Now()
	mimeType := "image/jpeg"
	uploadedBy := "user-1"
	updated := &domain.FileMetadata{
		ID:           id,
		OriginalName: newName,
		StoredName:   "abc123.jpg",
		Bucket:       "uploads",
		URL:          "https://example.com/abc123.jpg",
		Size:         1024,
		MimeType:     &mimeType,
		UploadedBy:   &uploadedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	repo.On("Update", ctx, id, req).Return(updated, nil)

	result, err := uc.UpdateFile(ctx, id, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newName, result.OriginalName)
	assert.Equal(t, id, result.ID)
	repo.AssertExpectations(t)
}

func TestUpdateFile_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := "file-456"
	newName := "renamed.jpg"
	req := domain.UpdateFileMetadataRequest{
		OriginalName: &newName,
	}

	repo.On("Update", ctx, id, req).Return(nil, errors.New("not found"))

	result, err := uc.UpdateFile(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}
