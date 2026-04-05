package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/content/generic/file/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---------------------------------------------------------------------------
// Mock types
// ---------------------------------------------------------------------------

type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error { return m.scanFunc(dest...) }

type mockRows struct {
	scanFunc func(dest ...any) error
}

func (m *mockRows) Scan(dest ...any) error                          { return m.scanFunc(dest...) }
func (m *mockRows) Close()                                          {}
func (m *mockRows) Err() error                                      { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                   { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription    { return nil }
func (m *mockRows) Next() bool                                      { return false }
func (m *mockRows) Values() ([]any, error)                          { return nil, nil }
func (m *mockRows) RawValues() [][]byte                             { return nil }
func (m *mockRows) Conn() *pgx.Conn                                { return nil }

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewFileWriteRepo(t *testing.T) {
	repo := NewFileWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewFileReadRepo(t *testing.T) {
	repo := NewFileReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanFileView (pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanFileView_Success(t *testing.T) {
	testID := uuid.New()
	uploaderID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		// readColumns: id, stored_name, original_name, mime_type, size, bucket, url, uploaded_by, created_at
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "abc123.pdf"
		*dest[2].(*string) = "document.pdf"
		*dest[3].(*string) = "application/pdf"
		*dest[4].(*int64) = 1024
		*dest[5].(*string) = "my-bucket"
		*dest[6].(*string) = "https://cdn.example.com/abc123.pdf"
		*dest[7].(**uuid.UUID) = &uploaderID
		*dest[8].(*time.Time) = now
		return nil
	}}

	v, err := scanFileView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != testID {
		t.Errorf("expected ID %v, got %v", testID, v.ID)
	}
	if v.Name != "abc123.pdf" {
		t.Errorf("expected name 'abc123.pdf', got %q", v.Name)
	}
	if v.OriginalName != "document.pdf" {
		t.Errorf("expected original_name 'document.pdf', got %q", v.OriginalName)
	}
	if v.MimeType != "application/pdf" {
		t.Errorf("expected mime_type 'application/pdf', got %q", v.MimeType)
	}
	if v.Size != 1024 {
		t.Errorf("expected size 1024, got %d", v.Size)
	}
	if v.Path != "my-bucket" {
		t.Errorf("expected path 'my-bucket', got %q", v.Path)
	}
	if v.URL != "https://cdn.example.com/abc123.pdf" {
		t.Errorf("expected URL, got %q", v.URL)
	}
	if v.UploadedBy == nil || *v.UploadedBy != uploaderID {
		t.Errorf("expected uploaded_by %v, got %v", uploaderID, v.UploadedBy)
	}
}

func TestScanFileView_NilUploadedBy(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "file.jpg"
		*dest[2].(*string) = "photo.jpg"
		*dest[3].(*string) = "image/jpeg"
		*dest[4].(*int64) = 2048
		*dest[5].(*string) = "bucket"
		*dest[6].(*string) = "https://cdn.example.com/file.jpg"
		*dest[7].(**uuid.UUID) = nil
		*dest[8].(*time.Time) = now
		return nil
	}}

	v, err := scanFileView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.UploadedBy != nil {
		t.Error("expected nil uploaded_by")
	}
}

func TestScanFileView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("scan error")
	}}

	_, err := scanFileView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanFileViewFromRows (pgx.Rows) — success
// ---------------------------------------------------------------------------

func TestScanFileViewFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "stored.png"
		*dest[2].(*string) = "original.png"
		*dest[3].(*string) = "image/png"
		*dest[4].(*int64) = 4096
		*dest[5].(*string) = "images"
		*dest[6].(*string) = "https://cdn.example.com/stored.png"
		*dest[7].(**uuid.UUID) = nil
		*dest[8].(*time.Time) = now
		return nil
	}}

	v, err := scanFileViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != testID {
		t.Errorf("expected ID %v, got %v", testID, v.ID)
	}
	if v.Size != 4096 {
		t.Errorf("expected size 4096, got %d", v.Size)
	}
}

func TestScanFileViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanFileViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.FileFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_NameOnly(t *testing.T) {
	name := "report"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.FileFilter{Name: &name})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_MimeTypeOnly(t *testing.T) {
	mt := "application/pdf"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.FileFilter{MimeType: &mt})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	name := "invoice"
	mt := "text/csv"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.FileFilter{Name: &name, MimeType: &mt})
	if len(result) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result))
	}
}
