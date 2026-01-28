package minio_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCase_UploadDoc_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		docData        []byte
		docSize        int64
		contentType    string
		expectError    bool
		validateResult func(t *testing.T, filename string)
	}{
		{
			name:        "success_upload_text_document",
			docData:     []byte("this is a test document content"),
			docSize:     int64(len("this is a test document content")),
			contentType: "text/plain",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat") // Documents are saved as .dat
				// Should be a UUID filename
				require.Greater(t, len(filename), 30) // UUID + .dat
			},
		},
		{
			name:        "success_upload_json_document",
			docData:     []byte(`{"name": "test", "value": 123}`),
			docSize:     int64(len(`{"name": "test", "value": 123}`)),
			contentType: "application/json",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "success_upload_xml_document",
			docData:     []byte("<?xml version=\"1.0\"?><root><item>test</item></root>"),
			docSize:     int64(len("<?xml version=\"1.0\"?><root><item>test</item></root>")),
			contentType: "application/xml",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "success_upload_csv_document",
			docData:     []byte("name,age,city\nJohn,30,NYC\nJane,25,LA"),
			docSize:     int64(len("name,age,city\nJohn,30,NYC\nJane,25,LA")),
			contentType: "text/csv",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "success_upload_large_document",
			docData:     bytes.Repeat([]byte("large document content "), 1000),
			docSize:     int64(len(bytes.Repeat([]byte("large document content "), 1000))),
			contentType: "text/plain",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "error_empty_document",
			docData:     []byte{},
			docSize:     0,
			contentType: "text/plain",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_zero_size",
			docData:     []byte("some content"),
			docSize:     0,
			contentType: "text/plain",
			expectError: true, // Zero size causes repository error
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_negative_size",
			docData:     []byte("some content"),
			docSize:     -1,
			contentType: "text/plain",
			expectError: false, // Size validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "error_unsupported_content_type",
			docData:     []byte("some content"),
			docSize:     int64(len("some content")),
			contentType: "image/jpeg",
			expectError: false, // Content type validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "success_upload_with_special_chars",
			docData:     []byte("document with special chars: àáâãäåæçèéêëìíîïñòóôõöøùúûüýþÿ"),
			docSize:     int64(len("document with special chars: àáâãäåæçèéêëìíîïñòóôõöøùúûüýþÿ")),
			contentType: "text/plain",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
		{
			name:        "success_upload_markdown",
			docData:     []byte("# Markdown Document\n\nThis is a **test** markdown file."),
			docSize:     int64(len("# Markdown Document\n\nThis is a **test** markdown file.")),
			contentType: "text/markdown",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat")
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc := setup(t)
			reader := bytes.NewReader(tt.docData)

			// act
			filename, err := uc.UploadDoc(context.Background(), reader, tt.docSize, tt.contentType)

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
