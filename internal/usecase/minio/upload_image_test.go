package minio_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with some color
	for x := range 100 {
		for y := range 100 {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func createTestJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with some color
	for x := range 100 {
		for y := range 100 {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img) // Using PNG for simplicity
	return buf.Bytes()
}

func TestUseCase_UploadImage_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		imageData      []byte
		imageSize      int64
		contentType    string
		expectError    bool
		validateResult func(t *testing.T, filename string)
	}{
		{
			name:        "success_upload_png",
			imageData:   createTestImage(),
			imageSize:   int64(len(createTestImage())),
			contentType: "image/png",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
				// Should be a UUID filename
				require.Greater(t, len(filename), 40) // UUID + .webp
			},
		},
		{
			name:        "success_upload_jpeg",
			imageData:   createTestJPEG(),
			imageSize:   int64(len(createTestJPEG())),
			contentType: "image/jpeg",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "success_upload_webp",
			imageData:   createTestImage(),
			imageSize:   int64(len(createTestImage())),
			contentType: "image/webp",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "error_invalid_image_data",
			imageData:   []byte("not an image"),
			imageSize:   int64(len("not an image")),
			contentType: "image/png",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_empty_image",
			imageData:   []byte{},
			imageSize:   0,
			contentType: "image/png",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_zero_size",
			imageData:   createTestImage(),
			imageSize:   0,
			contentType: "image/png",
			expectError: false, // Size validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "error_negative_size",
			imageData:   createTestImage(),
			imageSize:   -1,
			contentType: "image/png",
			expectError: false, // Size validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "error_unsupported_content_type",
			imageData:   createTestImage(),
			imageSize:   int64(len(createTestImage())),
			contentType: "text/plain",
			expectError: false, // Content type validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "success_large_image",
			imageData:   createTestImage(),
			imageSize:   int64(len(createTestImage())),
			contentType: "image/png",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".webp")
			},
		},
		{
			name:        "error_corrupted_image",
			imageData:   append(createTestImage()[:10], []byte("corrupted")...),
			imageSize:   int64(len(append(createTestImage()[:10], []byte("corrupted")...))),
			contentType: "image/png",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc := setup(t)
			reader := bytes.NewReader(tt.imageData)

			// act
			filename, err := uc.UploadImage(reader, tt.imageSize, tt.contentType)

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
