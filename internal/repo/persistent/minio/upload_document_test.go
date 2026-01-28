package minio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_UploadDocument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupDocument  func(t *testing.T) (string, int64, string, string)
		expectError    bool
		errorCheck     func(*testing.T, error)
		validateResult func(*testing.T, string, error)
	}{
		{
			name: "success_upload_pdf_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "fake pdf document content"
				size := int64(len(content))
				contentType := "application/pdf"
				expectedExt := ".pdf"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".pdf"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_docx_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "fake docx document content microsoft word"
				size := int64(len(content))
				contentType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
				expectedExt := ".docx"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".docx"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_xlsx_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "fake xlsx spreadsheet content microsoft excel"
				size := int64(len(content))
				contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
				expectedExt := ".xlsx"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".xlsx"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_pptx_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "fake pptx presentation content microsoft powerpoint"
				size := int64(len(content))
				contentType := "application/vnd.openxmlformats-officedocument.presentationml.presentation"
				expectedExt := ".pptx"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".pptx"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_txt_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "plain text document content"
				size := int64(len(content))
				contentType := "text/plain"
				expectedExt := ".txt"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".txt"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_csv_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "name,age,city\nJohn,25,NYC\nJane,30,LA"
				size := int64(len(content))
				contentType := "text/csv"
				expectedExt := ".csv"
				return content, size, contentType, expectedExt
			},
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
			name: "success_upload_rtf_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "{\\rtf1\\ansi\\ansicpg1252\\deff0\\deflang1033{\\fonttbl{\\f0\\fnil\\fcharset0 Calibri;}}\\viewkind4\\uc1\\pard\\sa200\\sl276\\slmult1\\f0\\fs22\\lang9 Rich Text Format Document\\par}"
				size := int64(len(content))
				contentType := "application/rtf"
				expectedExt := ".rtf"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".rtf"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "success_upload_large_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				// Simulate a large document
				content := strings.Repeat("Large document paragraph with lots of text content. ", 10000)
				size := int64(len(content))
				contentType := "application/pdf"
				expectedExt := ".pdf"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".pdf"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
		{
			name: "error_empty_content",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := ""
				size := int64(len(content))
				contentType := "application/pdf"
				expectedExt := ".pdf"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "error_invalid_content_type",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := "fake document content"
				size := int64(len(content))
				contentType := "image/jpeg" // Invalid for document
				expectedExt := ".jpg"
				return content, size, contentType, expectedExt
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success_upload_json_document",
			setupDocument: func(t *testing.T) (string, int64, string, string) {
				content := `{"title": "Document", "content": "This is a JSON document", "metadata": {"version": "1.0", "author": "test"}}`
				size := int64(len(content))
				contentType := "application/json"
				expectedExt := ".json"
				return content, size, contentType, expectedExt
			},
			expectError: false,
			validateResult: func(t *testing.T, fileName string, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, fileName)
				assert.True(t, strings.HasSuffix(fileName, ".json"))

				err = testRepo.ObjectExists(testCtx, fileName)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			content, size, contentType, _ := tt.setupDocument(t)
			reader := strings.NewReader(content)

			// act
			fileName, err := testRepo.UploadDocument(testCtx, reader, size, contentType)

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
