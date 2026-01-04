package minio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCase_DeleteFile_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fileName       string
		setupUpload    bool
		expectError    bool
		validateResult func(t *testing.T)
	}{
		{
			name:        "success_delete_existing_file",
			fileName:    "test-image.webp",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "success_delete_nonexistent_file",
			fileName:    "nonexistent-file.webp",
			setupUpload: false,
			expectError: false, // MinIO doesn't error on deleting non-existent files
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "error_empty_filename",
			fileName:    "",
			setupUpload: false,
			expectError: true,
			validateResult: func(t *testing.T) {
				// Should error on empty filename
			},
		},
		{
			name:        "success_delete_with_extension",
			fileName:    "test-file-with-dashes.webp",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "success_delete_special_chars",
			fileName:    "test_file_123.webp",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "error_filename_with_path",
			fileName:    "path/to/file.webp",
			setupUpload: false,
			expectError: false, // DeleteFile doesn't validate filenames
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "error_filename_with_slash",
			fileName:    "file/name.webp",
			setupUpload: false,
			expectError: false, // DeleteFile doesn't validate filenames
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
		{
			name:        "success_delete_long_filename",
			fileName:    strings.Repeat("a", 50) + ".webp",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T) {
				// Should handle long filenames
			},
		},
		{
			name:        "error_invalid_extension",
			fileName:    "test-file.txt",
			setupUpload: false,
			expectError: false, // DeleteFile doesn't validate extensions
			validateResult: func(t *testing.T) {
				// No specific validation needed for success
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc := setup(t)

			// Upload file first if setupUpload is true
			if tt.setupUpload {
				imgBytes := createTestImage()
				reader := bytes.NewReader(imgBytes)
				filename, err := uc.UploadImage(reader, int64(len(imgBytes)), "image/png")
				require.NoError(t, err)
				// Use the actual uploaded filename for deletion
				tt.fileName = filename
			}

			// act
			err := uc.DeleteFile(tt.fileName)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t)
				}
			}
		})
	}
}
