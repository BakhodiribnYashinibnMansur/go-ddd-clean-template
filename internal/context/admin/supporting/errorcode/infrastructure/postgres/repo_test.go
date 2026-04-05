package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/errorcode/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	// pgproto3 not needed

	"github.com/Masterminds/squirrel"
)

// ---------------------------------------------------------------------------
// Mocks
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

func TestNewErrorCodeWriteRepo(t *testing.T) {
	repo := NewErrorCodeWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewErrorCodeReadRepo(t *testing.T) {
	repo := NewErrorCodeReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// helpers to fill 13-column scan
// ---------------------------------------------------------------------------

func fillErrorCodeDest(dest []any, id uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "ERR_001"
	*dest[2].(*string) = "something failed"
	*dest[3].(*string) = "nimadir xato"
	*dest[4].(*string) = "что-то пошло не так"
	*dest[5].(*int) = 500
	*dest[6].(*string) = "server"
	*dest[7].(*string) = "critical"
	*dest[8].(*bool) = true
	*dest[9].(*int) = 30
	*dest[10].(*string) = "retry later"
	*dest[11].(*time.Time) = now
	*dest[12].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// scanErrorCodeView tests (read side)
// ---------------------------------------------------------------------------

func TestScanErrorCodeView_Success(t *testing.T) {
	id := domain.NewErrorCodeID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillErrorCodeDest(dest, id.UUID(), now)
		return nil
	}}

	v, err := scanErrorCodeView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Code != "ERR_001" {
		t.Errorf("Code = %q, want %q", v.Code, "ERR_001")
	}
	if v.HTTPStatus != 500 {
		t.Errorf("HTTPStatus = %d, want 500", v.HTTPStatus)
	}
	if !v.Retryable {
		t.Error("Retryable = false, want true")
	}
	if v.RetryAfter != 30 {
		t.Errorf("RetryAfter = %d, want 30", v.RetryAfter)
	}
}

func TestScanErrorCodeView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanErrorCodeView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanErrorCodeViewFromRows tests (read side)
// ---------------------------------------------------------------------------

func TestScanErrorCodeViewFromRows_Success(t *testing.T) {
	id := domain.NewErrorCodeID()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillErrorCodeDest(dest, id.UUID(), now)
		return nil
	}}

	v, err := scanErrorCodeViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanErrorCodeViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanErrorCodeViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanErrorCode tests (write side)
// ---------------------------------------------------------------------------

func TestScanErrorCode_Success(t *testing.T) {
	id := domain.NewErrorCodeID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillErrorCodeDest(dest, id.UUID(), now)
		return nil
	}}

	ec, err := scanErrorCode(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ec.TypedID() != id {
		t.Errorf("ID = %v, want %v", ec.ID(), id)
	}
	if ec.Code() != "ERR_001" {
		t.Errorf("Code = %q, want %q", ec.Code(), "ERR_001")
	}
	if ec.HTTPStatus() != 500 {
		t.Errorf("HTTPStatus = %d, want 500", ec.HTTPStatus())
	}
}

func TestScanErrorCode_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanErrorCode(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.ErrorCodeFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithCode(t *testing.T) {
	code := "ERR_001"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.ErrorCodeFilter{Code: &code})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithCategory(t *testing.T) {
	cat := "server"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.ErrorCodeFilter{Category: &cat})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithSeverity(t *testing.T) {
	sev := "critical"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.ErrorCodeFilter{Severity: &sev})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	code := "ERR_001"
	cat := "server"
	sev := "critical"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.ErrorCodeFilter{Code: &code, Category: &cat, Severity: &sev})
	if len(result) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(result))
	}
}
