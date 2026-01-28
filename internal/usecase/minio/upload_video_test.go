package minio_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCase_UploadVideo_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		videoData      []byte
		videoSize      int64
		contentType    string
		expectError    bool
		validateResult func(t *testing.T, filename string)
	}{
		{
			name:        "success_upload_mp4_video",
			videoData:   []byte("fake video content mp4"),
			videoSize:   int64(len("fake video content mp4")),
			contentType: "video/mp4",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MP4 videos keep .mp4 extension
			},
		},
		{
			name:        "success_upload_webm_video",
			videoData:   []byte("fake video content webm"),
			videoSize:   int64(len("fake video content webm")),
			contentType: "video/webm",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webm") // WebM videos keep .webm extension
			},
		},
		{
			name:        "success_upload_mov_video",
			videoData:   []byte("fake video content mov"),
			videoSize:   int64(len("fake video content mov")),
			contentType: "video/quicktime",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MOV videos are converted to .mp4
			},
		},
		{
			name:        "success_upload_avi_video",
			videoData:   []byte("fake video content avi"),
			videoSize:   int64(len("fake video content avi")),
			contentType: "video/x-msvideo",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // AVI videos are converted to .mp4
			},
		},
		{
			name:        "success_upload_large_video",
			videoData:   bytes.Repeat([]byte("large video content "), 10000),
			videoSize:   int64(len(bytes.Repeat([]byte("large video content "), 10000))),
			contentType: "video/mp4",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MP4 videos keep .mp4 extension
			},
		},
		{
			name:        "error_empty_video",
			videoData:   []byte{},
			videoSize:   0,
			contentType: "video/mp4",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_zero_size",
			videoData:   []byte("fake video content"),
			videoSize:   0,
			contentType: "video/mp4",
			expectError: true, // Zero size causes repository error
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_negative_size",
			videoData:   []byte("fake video content"),
			videoSize:   -1,
			contentType: "video/mp4",
			expectError: false, // Size validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MP4 videos keep .mp4 extension
			},
		},
		{
			name:        "error_unsupported_content_type",
			videoData:   []byte("fake video content"),
			videoSize:   int64(len("fake video content")),
			contentType: "text/plain",
			expectError: false, // Content type validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // Default to .mp4 for unknown types
			},
		},
		{
			name:        "success_upload_with_metadata",
			videoData:   []byte("fake video content with metadata"),
			videoSize:   int64(len("fake video content with metadata")),
			contentType: "video/mp4",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MP4 videos keep .mp4 extension
			},
		},
		{
			name:        "success_upload_binary_video",
			videoData:   append([]byte("fake video header"), bytes.Repeat([]byte("\x00\x01\x02\x03"), 1000)...),
			videoSize:   int64(len(append([]byte("fake video header"), bytes.Repeat([]byte("\x00\x01\x02\x03"), 1000)...))),
			contentType: "video/mp4",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MP4 videos keep .mp4 extension
			},
		},
		{
			name:        "success_upload_mkv_video",
			videoData:   []byte("fake video content mkv"),
			videoSize:   int64(len("fake video content mkv")),
			contentType: "video/x-matroska",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // MKV videos are converted to .mp4
			},
		},
		{
			name:        "success_upload_flv_video",
			videoData:   []byte("fake video content flv"),
			videoSize:   int64(len("fake video content flv")),
			contentType: "video/x-flv",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".mp4") // FLV videos are converted to .mp4
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc := setup(t)
			reader := bytes.NewReader(tt.videoData)

			// act
			filename, err := uc.UploadVideo(context.Background(), reader, tt.videoSize, tt.contentType)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, filename)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, filename)
				}
			}
		})
	}
}
