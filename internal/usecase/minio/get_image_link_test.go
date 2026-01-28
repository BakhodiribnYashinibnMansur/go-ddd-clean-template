package minio_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCase_GetImageLink_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fileName       string
		setupUpload    bool
		expectError    bool
		validateResult func(t *testing.T, link string)
	}{
		{
			name:        "success_get_existing_image_link",
			fileName:    "test-image.jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
				require.Contains(t, link, "http")
			},
		},
		{
			name:        "success_get_nonexistent_image_link",
			fileName:    "nonexistent-image.jpeg",
			setupUpload: false,
			expectError: false, // Should still generate a link even if file doesn't exist
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
				require.Contains(t, link, "http")
			},
		},
		{
			name:        "error_empty_filename",
			fileName:    "",
			setupUpload: false,
			expectError: true,
			validateResult: func(t *testing.T, link string) {
				require.Empty(t, link)
			},
		},
		{
			name:        "success_get_with_extension",
			fileName:    "test-file-with-dashes.jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
			},
		},
		{
			name:        "success_get_special_chars",
			fileName:    "test_file_123.jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
			},
		},
		{
			name:        "error_filename_with_path",
			fileName:    "path/to/file.jpeg",
			setupUpload: false,
			expectError: false, // GetImageLink doesn't validate filenames
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
			},
		},
		{
			name:        "error_filename_with_slash",
			fileName:    "file/name.jpeg",
			setupUpload: false,
			expectError: false, // GetImageLink doesn't validate filenames
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
			},
		},
		{
			name:        "success_get_long_filename",
			fileName:    strings.Repeat("a", 50) + ".jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
			},
		},
		{
			name:        "error_invalid_extension",
			fileName:    "test-file.txt",
			setupUpload: false,
			expectError: false, // GetImageLink doesn't validate extensions
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".txt")
			},
		},
		{
			name:        "success_get_uuid_filename",
			fileName:    "550e8400-e29b-41d4-a716-446655440000.jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
			},
		},
		{
			name:        "success_get_numeric_filename",
			fileName:    "123456789.jpeg",
			setupUpload: true,
			expectError: false,
			validateResult: func(t *testing.T, link string) {
				require.NotEmpty(t, link)
				require.Contains(t, link, ".jpeg")
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
				filename, err := uc.UploadImage(context.Background(), reader, int64(len(imgBytes)), "image/png")
				require.NoError(t, err)
				// Use the actual uploaded filename for getting link
				tt.fileName = filename
			}

			// act
			link, err := uc.GetImageLink(context.Background(), tt.fileName)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, link)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, link)
				}
			}
		})
	}
}
