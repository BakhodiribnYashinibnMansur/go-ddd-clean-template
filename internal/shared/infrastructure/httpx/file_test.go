package httpx

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestFileTransfer_Success(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-download.txt")
	content := []byte("hello file transfer")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/download", nil)

	err := FileTransfer(c, tmpFile, "text/plain")
	if err != nil {
		t.Fatalf("FileTransfer returned error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "hello file transfer" {
		t.Errorf("expected body 'hello file transfer', got %q", w.Body.String())
	}

	cd := w.Header().Get(HeaderContentDisposition)
	expected := AttachmentPrefix + "test-download.txt"
	if cd != expected {
		t.Errorf("expected Content-Disposition %q, got %q", expected, cd)
	}

	desc := w.Header().Get(HeaderContentDescription)
	if desc != FileTransferDescription {
		t.Errorf("expected Content-Description %q, got %q", FileTransferDescription, desc)
	}
}

func TestFileTransfer_FileNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/download", nil)

	err := FileTransfer(c, "/nonexistent/file.txt", "text/plain")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestDownloadFile_Success(t *testing.T) {
	// Create a temp file relative to CurrentDir
	tmpFile := "test-dl-" + t.Name() + ".txt"
	fullPath := CurrentDir + tmpFile
	content := []byte("download content")
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(fullPath)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/download", nil)

	err := DownloadFile(c, tmpFile)
	if err != nil {
		t.Fatalf("DownloadFile returned error: %v", err)
	}

	if w.Body.String() != "download content" {
		t.Errorf("expected body 'download content', got %q", w.Body.String())
	}

	cd := w.Header().Get(HeaderContentDisposition)
	expected := AttachmentPrefix + tmpFile
	if cd != expected {
		t.Errorf("expected Content-Disposition %q, got %q", expected, cd)
	}
}

func TestDownloadFile_FileNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/download", nil)

	err := DownloadFile(c, "/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
