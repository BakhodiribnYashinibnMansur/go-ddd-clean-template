package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gct/internal/domain"
)

func TestRepo_GetFileURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFiles     func(t *testing.T) []domain.File
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, []domain.File, error)
	}{
		{
			name: "success_get_multiple_image_urls",
			setupFiles: func(t *testing.T) []domain.File {
				content := "image content for multiple urls test"
				reader := strings.NewReader(content)
				f1, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				reader2 := strings.NewReader(content)
				f2, err := testRepo.UploadImage(testCtx, reader2, int64(len(content)), "image/jpeg")
				require.NoError(t, err)

				return []domain.File{
					{Name: f1},
					{Name: f2},
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 2)

				assert.NotEmpty(t, files[0].Link)
				assert.Contains(t, files[0].Link, files[0].Name)

				assert.NotEmpty(t, files[1].Link)
				assert.Contains(t, files[1].Link, files[1].Name)
			},
		},
		{
			name: "success_get_mixed_file_type_urls",
			setupFiles: func(t *testing.T) []domain.File {
				// Upload image
				imageContent := "image content"
				reader := strings.NewReader(imageContent)
				f1, err := testRepo.UploadImage(testCtx, reader, int64(len(imageContent)), "image/png")
				require.NoError(t, err)

				// Upload document
				docContent := "document content"
				reader2 := strings.NewReader(docContent)
				f2, err := testRepo.UploadDocument(testCtx, reader2, int64(len(docContent)), "application/pdf")
				require.NoError(t, err)

				// Upload video
				videoContent := "video content"
				reader3 := strings.NewReader(videoContent)
				f3, err := testRepo.UploadVideo(testCtx, reader3, int64(len(videoContent)), "video/mp4")
				require.NoError(t, err)

				return []domain.File{
					{Name: f1},
					{Name: f2},
					{Name: f3},
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 3)

				for _, file := range files {
					assert.NotEmpty(t, file.Link)
					assert.Contains(t, file.Link, file.Name)
					assert.Contains(t, file.Link, testRepo.config.Bucket)
				}
			},
		},
		{
			name: "success_with_empty_filename",
			setupFiles: func(t *testing.T) []domain.File {
				content := "content"
				reader := strings.NewReader(content)
				f1, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				return []domain.File{
					{Name: f1},
					{Name: ""}, // Empty name should result in empty link
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 2)

				assert.NotEmpty(t, files[0].Link)
				assert.Contains(t, files[0].Link, files[0].Name)

				assert.Empty(t, files[1].Link) // Should be empty link for empty name
			},
		},
		{
			name: "success_with_nonexistent_files",
			setupFiles: func(t *testing.T) []domain.File {
				return []domain.File{
					{Name: "nonexistent1.png"},
					{Name: "nonexistent2.pdf"},
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 2)

				// Nonexistent files should have empty links
				assert.Empty(t, files[0].Link)
				assert.Empty(t, files[1].Link)
			},
		},
		{
			name: "success_mixed_existing_and_nonexistent",
			setupFiles: func(t *testing.T) []domain.File {
				content := "content"
				reader := strings.NewReader(content)
				f1, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				return []domain.File{
					{Name: f1},
					{Name: "nonexistent.png"},
					{Name: ""}, // Empty name
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 3)

				assert.NotEmpty(t, files[0].Link)
				assert.Contains(t, files[0].Link, files[0].Name)

				assert.Empty(t, files[1].Link) // Nonexistent file
				assert.Empty(t, files[2].Link) // Empty name
			},
		},
		{
			name: "success_empty_file_list",
			setupFiles: func(t *testing.T) []domain.File {
				return []domain.File{}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 0)
			},
		},
		{
			name: "success_large_file_list",
			setupFiles: func(t *testing.T) []domain.File {
				var files []domain.File
				content := "content for large list test"

				for i := 0; i < 10; i++ {
					reader := strings.NewReader(content)
					f, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
					require.NoError(t, err)
					files = append(files, domain.File{Name: f})
				}

				return files
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 10)

				for _, file := range files {
					assert.NotEmpty(t, file.Link)
					assert.Contains(t, file.Link, file.Name)
					assert.Contains(t, file.Link, testRepo.config.Bucket)
				}
			},
		},
		{
			name: "success_with_special_characters",
			setupFiles: func(t *testing.T) []domain.File {
				content := "special content"
				reader := strings.NewReader(content)
				f1, err := testRepo.UploadImage(testCtx, reader, int64(len(content)), "image/png")
				require.NoError(t, err)

				return []domain.File{
					{Name: f1},
					{Name: "file with spaces.png"},
					{Name: "file-with-dashes.jpg"},
					{Name: "file_with_underscores.gif"},
				}
			},
			expectError: false,
			validateResult: func(t *testing.T, files []domain.File, err error) {
				require.NoError(t, err)
				assert.Len(t, files, 4)

				// First file exists and should have a URL
				assert.NotEmpty(t, files[0].Link)
				assert.Contains(t, files[0].Link, files[0].Name)

				// Other files don't exist, should have empty links
				assert.Empty(t, files[1].Link)
				assert.Empty(t, files[2].Link)
				assert.Empty(t, files[3].Link)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			files := tt.setupFiles(t)

			// act
			urls, err := testRepo.GetFileURLs(testCtx, files)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, urls, err)
				}
			}
		})
	}
}
