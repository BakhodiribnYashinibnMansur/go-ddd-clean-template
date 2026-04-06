package postgres

import (
	"errors"
	"testing"
	"time"

	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"

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

func (m *mockRows) Scan(dest ...any) error                       { return m.scanFunc(dest...) }
func (m *mockRows) Close()                                       {}
func (m *mockRows) Err() error                                   { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool                                   { return false }
func (m *mockRows) Values() ([]any, error)                       { return nil, nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewDataExportWriteRepo(t *testing.T) {
	repo := NewDataExportWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewDataExportReadRepo(t *testing.T) {
	repo := NewDataExportReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanDataExport (write_repo.go — pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanDataExport_Success(t *testing.T) {
	testID := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		// writeColumns: id, type, status, file_url, created_by, created_at, completed_at
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "users"
		*dest[2].(*string) = "COMPLETED"
		*dest[3].(*string) = "https://example.com/file.csv"
		*dest[4].(**uuid.UUID) = &userID
		*dest[5].(*time.Time) = now
		*dest[6].(**time.Time) = nil
		return nil
	}}

	de, err := scanDataExport(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if de == nil {
		t.Fatal("expected non-nil DataExport")
	}
	if de.ID() != testID {
		t.Errorf("expected ID %v, got %v", testID, de.ID())
	}
	if de.DataType() != "users" {
		t.Errorf("expected data type 'users', got %q", de.DataType())
	}
	if de.Status() != "COMPLETED" {
		t.Errorf("expected status 'COMPLETED', got %q", de.Status())
	}
}

func TestScanDataExport_NilCreatedBy(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "orders"
		*dest[2].(*string) = "PENDING"
		*dest[3].(*string) = ""
		*dest[4].(**uuid.UUID) = nil
		*dest[5].(*time.Time) = now
		*dest[6].(**time.Time) = nil
		return nil
	}}

	de, err := scanDataExport(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if de.UserID() != uuid.Nil {
		t.Errorf("expected nil user ID, got %v", de.UserID())
	}
}

func TestScanDataExport_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("scan error")
	}}

	_, err := scanDataExport(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanDataExportView (read_repo.go — pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanDataExportView_Success(t *testing.T) {
	testID := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "logs"
		*dest[2].(*string) = "PROCESSING"
		*dest[3].(*string) = "https://example.com/export.csv"
		*dest[4].(**uuid.UUID) = &userID
		*dest[5].(*time.Time) = now
		*dest[6].(**time.Time) = nil
		return nil
	}}

	v, err := scanDataExportView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != exportentity.DataExportID(testID) {
		t.Errorf("expected ID %v, got %v", testID, v.ID)
	}
	if v.DataType != "logs" {
		t.Errorf("expected data type 'logs', got %q", v.DataType)
	}
	if v.UserID != userID {
		t.Errorf("expected user_id %v, got %v", userID, v.UserID)
	}
	if v.FileURL == nil || *v.FileURL != "https://example.com/export.csv" {
		t.Errorf("expected file_url set, got %v", v.FileURL)
	}
}

func TestScanDataExportView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("view scan error")
	}}

	_, err := scanDataExportView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanDataExportViewFromRows (pgx.Rows) — success
// ---------------------------------------------------------------------------

func TestScanDataExportViewFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "events"
		*dest[2].(*string) = "FAILED"
		*dest[3].(*string) = ""
		*dest[4].(**uuid.UUID) = nil
		*dest[5].(*time.Time) = now
		*dest[6].(**time.Time) = nil
		return nil
	}}

	v, err := scanDataExportViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != "FAILED" {
		t.Errorf("expected status 'FAILED', got %q", v.Status)
	}
	if v.FileURL != nil {
		t.Error("expected nil file_url for empty string")
	}
}

func TestScanDataExportViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanDataExportViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, exportrepo.DataExportFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_UserIDOnly(t *testing.T) {
	uid := uuid.New()
	conds := squirrel.And{}
	result := applyFilters(conds, exportrepo.DataExportFilter{UserID: &uid})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_DataTypeOnly(t *testing.T) {
	dt := "users"
	conds := squirrel.And{}
	result := applyFilters(conds, exportrepo.DataExportFilter{DataType: &dt})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_StatusOnly(t *testing.T) {
	s := "COMPLETED"
	conds := squirrel.And{}
	result := applyFilters(conds, exportrepo.DataExportFilter{Status: &s})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	uid := uuid.New()
	dt := "orders"
	s := "PENDING"
	conds := squirrel.And{}
	result := applyFilters(conds, exportrepo.DataExportFilter{UserID: &uid, DataType: &dt, Status: &s})
	if len(result) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(result))
	}
}
