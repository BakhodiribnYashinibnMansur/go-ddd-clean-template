package postgres

import (
	"errors"
	"testing"
	"time"

	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"

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
// helpers for pointer values
// ---------------------------------------------------------------------------

func strPtr(s string) *string       { return &s }
func uuidPtr(u uuid.UUID) *uuid.UUID { return &u }
func timePtr(t time.Time) *time.Time { return &t }
func boolPtr(b bool) *bool           { return &b }

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewSystemErrorWriteRepo(t *testing.T) {
	repo := NewSystemErrorWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewSystemErrorReadRepo(t *testing.T) {
	repo := NewSystemErrorReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// mock scan filler — fills the 16-column systemerror scan dest
// ---------------------------------------------------------------------------

func fillSystemErrorDest(dest []any, id uuid.UUID, now time.Time) {
	// writeColumns: id, code, message, stack_trace,
	//   severity, service_name, request_id, user_id,
	//   ip_address, path, method,
	//   is_resolved, resolved_at, resolved_by, created_at
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "ERR_500"
	*dest[2].(*string) = "Internal Server Error"
	*dest[3].(**string) = strPtr("goroutine 1 [running]")
	*dest[4].(*string) = "critical"
	*dest[5].(**string) = strPtr("api-service")
	*dest[6].(**uuid.UUID) = uuidPtr(uuid.New())
	*dest[7].(**uuid.UUID) = uuidPtr(uuid.New())
	*dest[8].(**string) = strPtr("192.168.1.1")
	*dest[9].(**string) = strPtr("/api/users")
	*dest[10].(**string) = strPtr("POST")
	*dest[11].(*bool) = false
	*dest[12].(**time.Time) = nil
	*dest[13].(**uuid.UUID) = nil
	*dest[14].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// scanSystemError (pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanSystemError_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillSystemErrorDest(dest, testID, now)
		return nil
	}}

	se, err := scanSystemError(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if se == nil {
		t.Fatal("expected non-nil SystemError")
	}
	if se.ID() != testID {
		t.Errorf("expected ID %v, got %v", testID, se.ID())
	}
	if se.Code() != "ERR_500" {
		t.Errorf("expected code 'ERR_500', got %q", se.Code())
	}
	if se.Message() != "Internal Server Error" {
		t.Errorf("expected message 'Internal Server Error', got %q", se.Message())
	}
	if se.Severity() != "critical" {
		t.Errorf("expected severity 'critical', got %q", se.Severity())
	}
	if se.IsResolved() {
		t.Error("expected is_resolved false")
	}
}

func TestScanSystemError_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("scan error")
	}}

	_, err := scanSystemError(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanSystemErrorFromRows (pgx.Rows) — success
// ---------------------------------------------------------------------------

func TestScanSystemErrorFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillSystemErrorDest(dest, testID, now)
		return nil
	}}

	se, err := scanSystemErrorFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if se.Code() != "ERR_500" {
		t.Errorf("expected code 'ERR_500', got %q", se.Code())
	}
}

func TestScanSystemErrorFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanSystemErrorFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanView / scanViewFromRows (read_repo.go)
// ---------------------------------------------------------------------------

func fillSystemErrorViewDest(dest []any, id uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "ERR_404"
	*dest[2].(*string) = "Not Found"
	*dest[3].(**string) = nil
	*dest[4].(*string) = "warning"
	*dest[5].(**string) = strPtr("web-service")
	*dest[6].(**uuid.UUID) = nil
	*dest[7].(**uuid.UUID) = nil
	*dest[8].(**string) = strPtr("10.0.0.1")
	*dest[9].(**string) = strPtr("/api/items/999")
	*dest[10].(**string) = strPtr("GET")
	*dest[11].(*bool) = true
	resolvedTime := now.Add(-time.Hour)
	*dest[12].(**time.Time) = &resolvedTime
	resolverID := uuid.New()
	*dest[13].(**uuid.UUID) = &resolverID
	*dest[14].(*time.Time) = now
}

func TestScanView_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillSystemErrorViewDest(dest, testID, now)
		return nil
	}}

	v, err := scanView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != syserrentity.SystemErrorID(testID) {
		t.Errorf("expected ID %v, got %v", testID, v.ID)
	}
	if v.Code != "ERR_404" {
		t.Errorf("expected code 'ERR_404', got %q", v.Code)
	}
	if !v.IsResolved {
		t.Error("expected is_resolved true")
	}
	if v.ResolvedAt == nil {
		t.Error("expected non-nil resolved_at")
	}
	if v.ResolvedBy == nil {
		t.Error("expected non-nil resolved_by")
	}
}

func TestScanView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("view scan error")
	}}

	_, err := scanView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestScanViewFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillSystemErrorViewDest(dest, testID, now)
		return nil
	}}

	v, err := scanViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Code != "ERR_404" {
		t.Errorf("expected code 'ERR_404', got %q", v.Code)
	}
}

func TestScanViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_CodeOnly(t *testing.T) {
	code := "ERR_500"
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{Code: &code})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_SeverityOnly(t *testing.T) {
	sev := "critical"
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{Severity: &sev})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_IsResolvedOnly(t *testing.T) {
	resolved := false
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{IsResolved: &resolved})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_FromDateOnly(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour)
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{FromDate: &from})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_ToDateOnly(t *testing.T) {
	to := time.Now()
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{ToDate: &to})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_RequestIDOnly(t *testing.T) {
	rid := uuid.New()
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{RequestID: &rid})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_UserIDOnly(t *testing.T) {
	uid := uuid.New()
	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{UserID: &uid})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	code := "ERR_500"
	sev := "critical"
	resolved := true
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	rid := uuid.New()
	uid := uuid.New()

	conds := squirrel.And{}
	result := applyFilters(conds, syserrrepo.SystemErrorFilter{
		Code:       &code,
		Severity:   &sev,
		IsResolved: &resolved,
		FromDate:   &from,
		ToDate:     &to,
		RequestID:  &rid,
		UserID:     &uid,
	})
	if len(result) != 7 {
		t.Errorf("expected 7 conditions, got %d", len(result))
	}
}
