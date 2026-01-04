package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_UploadImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupImage     func(t *testing.T) (string, int64, string, string)
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, string, error)
	}{
		{
			name: "success_upload_jpeg_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := "fake jpeg image content"
				size := int64(len(content))
				contentType := "image/jpeg"
				expectedExt := ".jpeg"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".jpeg") || strings.HasSuffix(fileName, ".jpg"), "filename should have .jpeg or .jpg extension")

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_png_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := "fake png image content with transparency"
				size := int64(len(content))
				contentType := "image/png"
				expectedExt := ".png"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".png"), "filename should have .png extension")

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_gif_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := "fake gif image content animated"
				size := int64(len(content))
				contentType := "image/gif"
				expectedExt := ".gif"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".gif"), "filename should have .gif extension")

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_webp_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := "fake webp image content modern format"
				size := int64(len(content))
				contentType := "image/webp"
				expectedExt := ".webp"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".webp"), "filename should have .webp extension")

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_large_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				// Simulate a large image
				content := strings.Repeat("large image pixel data", 10000)
				size := int64(len(content))
				contentType := "image/jpeg"
				expectedExt := ".jpeg"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".jpeg") || strings.HasSuffix(fileName, ".jpg"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "error_empty_content",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := ""
				size := int64(len(content))
				contentType := "image/jpeg"
				expectedExt := ".jpeg"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "error_invalid_content_type",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := "fake image content"
				size := int64(len(content))
				contentType := "text/plain" // Invalid for image
				expectedExt := ".txt"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_upload_svg_image",
			setupImage: func(t *testing.T) (string, int64, string, string) {
				content := `<svg><circle cx="50" cy="50" r="40" stroke="black" stroke-width="3" fill="red" /></svg>`
				size := int64(len(content))
				contentType := "image/svg+xml"
				expectedExt := ".svg"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".svg"), "filename should have .svg extension")

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			content, size, contentType, _ := tt.setupImage(t)
			reader := strings.NewReader(content)

			// act
			fileName, err := testRepo.UploadImage(testCtx, reader, size, contentType)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, fileName, err)
				}
			}
		})
	}
}
