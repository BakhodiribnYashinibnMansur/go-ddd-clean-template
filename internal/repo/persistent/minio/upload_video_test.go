package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_UploadVideo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupVideo     func(t *testing.T) (string, int64, string, string)
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, string, error)
	}{
		{
			name: "success_upload_mp4_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake mp4 video content"
				size := int64(len(content))
				contentType := "video/mp4"
				expectedExt := ".mp4"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".mp4"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_avi_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake avi video content classic format"
				size := int64(len(content))
				contentType := "video/avi"
				expectedExt := ".avi"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".avi"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_mov_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake mov video content apple format"
				size := int64(len(content))
				contentType := "video/quicktime"
				expectedExt := ".mov"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".mov"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_webm_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake webm video content web format"
				size := int64(len(content))
				contentType := "video/webm"
				expectedExt := ".webm"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".webm"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_large_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				// Simulate a large video file
				content := strings.Repeat("large video frame data chunk", 50000)
				size := int64(len(content))
				contentType := "video/mp4"
				expectedExt := ".mp4"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".mp4"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_mkv_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake mkv video content matroska format"
				size := int64(len(content))
				contentType := "video/x-matroska"
				expectedExt := ".mkv"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".mkv"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "error_empty_content",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := ""
				size := int64(len(content))
				contentType := "video/mp4"
				expectedExt := ".mp4"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "error_invalid_content_type",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake video content"
				size := int64(len(content))
				contentType := "text/plain" // Invalid for video
				expectedExt := ".txt"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_upload_3gp_video",
			setupVideo: func(t *testing.T) (string, int64, string, string) {
				content := "fake 3gp video content mobile format"
				size := int64(len(content))
				contentType := "video/3gpp"
				expectedExt := ".3gp"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".3gp"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			content, size, contentType, _ := tt.setupVideo(t)
			reader := strings.NewReader(content)

			// act
			fileName, err := testRepo.UploadVideo(testCtx, reader, size, contentType)

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
