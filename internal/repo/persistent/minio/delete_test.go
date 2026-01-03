package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_DeleteFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFile     func(t *testing.T) string
		expectError   bool
		errorCheck    func(*testing.T, error)
		validateState func(*testing.T, string)
	}{
		{
			name: "success_delete_existing_image",
			setupFile: func(t *testing.T) string {
				content := "to delete"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				// Verify file exists before delete
				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string) {
				// File should no longer exist after delete
				err := testRepo.ObjectExists(testCtx, fileName)
				assert.Error(t, err) // Should error now
			},
		},
		{
			name: "success_delete_uploaded_document",
			setupFile: func(t *testing.T) string {
				content := "document to delete"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadDocument(testCtx, reader, int64(len(content)), "application/pdf")
				require.NoError(t, err)

				// Verify file exists before delete
				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string) {
				err := testRepo.ObjectExists(testCtx, fileName)
				assert.Error(t, err)
			},
		},
		{
			name: "success_delete_uploaded_video",
			setupFile: func(t *testing.T) string {
				content := "video content to delete"
				reader := strings.NewReader(content)
				fileName, err := testRepo.UploadVideo(testCtx, reader, int64(len(content)), "video/mp4")
				require.NoError(t, err)

				// Verify file exists before delete
				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
				return fileName
			},
			expectError: false,
			validateState: func(t *testing.T, fileName string) {
				err := testRepo.ObjectExists(testCtx, fileName)
				assert.Error(t, err)
			},
		},
		{
			name: "error_delete_nonexistent_file",
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
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			fileName := tt.setupFile(t)

			// act
			err := testRepo.DeleteFile(testCtx, fileName)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateState != nil {
					tt.validateState(t, fileName)
				}
			}
		})
	}
}
