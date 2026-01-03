package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_ObjectExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFile     func(t *testing.T) string
		expectError   bool
		errorCheck    func(*testing.T, error)
		validateState func(*testing.T, string, error)
	}{
		{
			name: "success_existing_image",
			setupFile: func(t *testing.T) string {
				content := "exists test image"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success_existing_document",
			setupFile: func(t *testing.T) string {
				content := "exists test document"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "application/pdf")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success_existing_video",
			setupFile: func(t *testing.T) string {
				content := "exists test video"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadVideo(testCtx, reader, int64(len(content)), "video/mp4")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success_existing_file",
			setupFile: func(t *testing.T) string {
				content := "exists test file"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "error_nonexistent_file",
			setupFile: func(t *testing.T) string {
				return "non-existent-file.txt"
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "error_empty_filename",
			setupFile: func(t *testing.T) string {
				return ""
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "error_deleted_file",
			setupFile: func(t *testing.T) string {
				content := "file to be deleted"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				// Delete the file first
				err = testRepo.DeleteFile(testCtx, fileName)
				require.NoError(t, err)

				return fileName
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_large_file_exists",
			setupFile: func(t *testing.T) string {
				content := strings.Repeat("large file content for existence test ", 1000)
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "application/pdf")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "error_invalid_filename_format",
			setupFile: func(t *testing.T) string {
				return "../invalid/path/file.txt"
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_unicode_filename_exists",
			setupFile: func(t *testing.T) string {
				content := "unicode content тест"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "text/plain")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "error_special_characters_nonexistent",
			setupFile: func(t *testing.T) string {
				return "file with spaces & symbols!.txt"
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			fileName := tt.setupFile(t)

			// act
			err := testRepo.ObjectExists(testCtx, fileName)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateState != nil {
					tt.validateState(t, fileName, err)
				}
			}
		})
	}
}
