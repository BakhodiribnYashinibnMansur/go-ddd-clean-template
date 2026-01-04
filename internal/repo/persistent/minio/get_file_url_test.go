package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_GetFileURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFile      func(t *testing.T) string
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, string, string, error)
	}{
		{
			name: "success_get_image_url",
			setupFile: func(t *testing.T) string {
				content := "image content for url test"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
			},
		},
		{
			name: "success_get_document_url",
			setupFile: func(t *testing.T) string {
				content := "document content for url test"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "application/pdf")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
			},
		},
		{
			name: "success_get_video_url",
			setupFile: func(t *testing.T) string {
				content := "video content for url test"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadVideo(testCtx, reader, int64(len(content)), "video/mp4")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
			},
		},
		{
			name: "success_get_file_url_with_special_chars",
			setupFile: func(t *testing.T) string {
				content := "file content with special chars"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
			},
		},
		{
			name: "error_nonexistent_file",
			setupFile: func(t *testing.T) string {
				return "nonexistent-file.txt"
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
			name: "success_get_large_file_url",
			setupFile: func(t *testing.T) string {
				content := strings.Repeat("large file content for url generation test ", 1000)
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "application/pdf")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
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
			name: "success_get_url_with_unicode_filename",
			setupFile: func(t *testing.T) string {
				content := "unicode file content тест"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "text/plain")
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName, url string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, fileName)
				assert.Contains(t, url, testRepo.config.Bucket)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			fileName := tt.setupFile(t)

			// act
			url, err := testRepo.GetFileURL(testCtx, fileName)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, fileName, url, err)
				}
			}
		})
	}
}
