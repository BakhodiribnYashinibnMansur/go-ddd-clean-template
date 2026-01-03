package minio

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_UploadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFile      func(t *testing.T) (string, string)
		contentType    string
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, string, error)
	}{
		{
			name: "success_upload_text_file",
			setupFile: func(t *testing.T) (string, string) {
				content := "Hello MinIO"
				tmpFile, err := os.CreateTemp("", "test-file-*.txt")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name(), ".txt"
			},
			contentType: "text/plain",
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".txt"))

				// Verify file exists
				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_json_file",
			setupFile: func(t *testing.T) (string, string) {
				content := `{"name": "test", "value": 123}`
				tmpFile, err := os.CreateTemp("", "test-file-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name(), ".json"
			},
			contentType: "application/json",
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".json"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_csv_file",
			setupFile: func(t *testing.T) (string, string) {
				content := "name,age,city\nJohn,25,NYC\nJane,30,LA"
				tmpFile, err := os.CreateTemp("", "test-file-*.csv")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name(), ".csv"
			},
			contentType: "text/csv",
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".csv"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "error_file_not_found",
			setupFile: func(t *testing.T) (string, string) {
				return "/tmp/nonexistent.txt", ".txt"
			},
			contentType: "text/plain",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no such file")
			},
		},
		{
			name: "error_empty_file_path",
			setupFile: func(t *testing.T) (string, string) {
				return "", ".txt"
			},
			contentType: "text/plain",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_upload_large_file",
			setupFile: func(t *testing.T) (string, string) {
				// Create a larger file
				content := strings.Repeat("Large file content line\n", 1000)
				tmpFile, err := os.CreateTemp("", "large-file-*.txt")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name(), ".txt"
			},
			contentType: "text/plain",
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".txt"))

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
			filePath, _ := tt.setupFile(t)

			// act
			fileName, err := testRepo.UploadFile(testCtx, filePath, tt.contentType)

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
