package minio_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCase_UploadPDF_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pdfData        []byte
		pdfSize        int64
		contentType    string
		expectError    bool
		validateResult func(t *testing.T, filename string)
	}{
		{
			name:        "success_upload_minimal_pdf",
			pdfData:     []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f\n0000000010 00000 n\n0000000053 00000 n\n0000000145 00000 n\ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n164\n%%EOF"),
			pdfSize:     int64(len("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f\n0000000010 00000 n\n0000000053 00000 n\n0000000145 00000 n\ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n164\n%%EOF")),
			contentType: "application/pdf",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "success_upload_simple_pdf",
			pdfData:     []byte("%PDF-1.4\n..."),
			pdfSize:     int64(len("%PDF-1.4\n...")),
			contentType: "application/pdf",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "success_upload_large_pdf",
			pdfData:     bytes.Repeat([]byte("%PDF-1.4\n"), 1000),
			pdfSize:     int64(len(bytes.Repeat([]byte("%PDF-1.4\n"), 1000))),
			contentType: "application/pdf",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "error_empty_pdf",
			pdfData:     []byte{},
			pdfSize:     0,
			contentType: "application/pdf",
			expectError: true,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_zero_size",
			pdfData:     []byte("%PDF-1.4\n..."),
			pdfSize:     0,
			contentType: "application/pdf",
			expectError: true, // Zero size causes repository error
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.Empty(t, filename)
			},
		},
		{
			name:        "error_negative_size",
			pdfData:     []byte("%PDF-1.4\n..."),
			pdfSize:     -1,
			contentType: "application/pdf",
			expectError: false, // Size validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "error_invalid_content_type",
			pdfData:     []byte("%PDF-1.4\n..."),
			pdfSize:     int64(len("%PDF-1.4\n...")),
			contentType: "text/plain",
			expectError: false, // Content type validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".dat") // Non-PDF content types are saved as .dat
			},
		},
		{
			name:        "error_not_pdf_format",
			pdfData:     []byte("this is not a pdf file"),
			pdfSize:     int64(len("this is not a pdf file")),
			contentType: "application/pdf",
			expectError: false, // PDF format validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "error_pdf_without_header",
			pdfData:     []byte("some random content without pdf header"),
			pdfSize:     int64(len("some random content without pdf header")),
			contentType: "application/pdf",
			expectError: false, // PDF header validation is not performed
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "success_pdf_with_metadata",
			pdfData:     []byte("%PDF-1.7\n%âãÏÓ\n1 0 obj\n<<\n/Creator (Test Creator)\n/Producer (Test Producer)\n/CreationDate (D:20231231120000+00'00')\n>>\nendobj\n..."),
			pdfSize:     int64(len("%PDF-1.7\n%âãÏÓ\n1 0 obj\n<<\n/Creator (Test Creator)\n/Producer (Test Producer)\n/CreationDate (D:20231231120000+00'00')\n>>\nendobj\n...")),
			contentType: "application/pdf",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
		{
			name:        "success_pdf_binary_content",
			pdfData:     append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("\x00\x01\x02\x03"), 100)...),
			pdfSize:     int64(len(append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("\x00\x01\x02\x03"), 100)...))),
			contentType: "application/pdf",
			expectError: false,
			validateResult: func(t *testing.T, filename string) {
				t.Helper()
				require.NotEmpty(t, filename)
				require.Contains(t, filename, ".pdf") // PDFs keep .pdf extension
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc := setup(t)
			reader := bytes.NewReader(tt.pdfData)

			// act
			filename, err := uc.UploadPDF(reader, tt.pdfSize, tt.contentType)

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
