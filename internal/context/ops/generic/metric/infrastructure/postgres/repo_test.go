package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/generic/metric/domain"

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

func (m *mockRows) Scan(dest ...any) error                        { return m.scanFunc(dest...) }
func (m *mockRows) Close()                                        {}
func (m *mockRows) Err() error                                    { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                 { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool                                    { return false }
func (m *mockRows) Values() ([]any, error)                        { return nil, nil }
func (m *mockRows) RawValues() [][]byte                           { return nil }
func (m *mockRows) Conn() *pgx.Conn                               { return nil }

// ---------------------------------------------------------------------------
// helpers — 6 columns: id, name, latency_ms, is_panic, panic_error, created_at
// ---------------------------------------------------------------------------

func fillMetricDest(dest []any, id uuid.UUID, now time.Time) {
	panicErr := "something panicked"
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "handleRequest"
	*dest[2].(*float64) = 42.5
	*dest[3].(*bool) = true
	*dest[4].(**string) = &panicErr
	*dest[5].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewMetricWriteRepo(t *testing.T) {
	repo := NewMetricWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewMetricReadRepo(t *testing.T) {
	repo := NewMetricReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanMetricView tests (read side — takes pgx.Rows)
// ---------------------------------------------------------------------------

func TestScanMetricView_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillMetricDest(dest, id, now)
		return nil
	}}

	v, err := scanMetricView(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Name != "handleRequest" {
		t.Errorf("Name = %q, want %q", v.Name, "handleRequest")
	}
	if v.LatencyMs != 42.5 {
		t.Errorf("LatencyMs = %f, want 42.5", v.LatencyMs)
	}
	if !v.IsPanic {
		t.Error("IsPanic = false, want true")
	}
	if v.PanicError == nil || *v.PanicError != "something panicked" {
		t.Errorf("PanicError = %v, want 'something panicked'", v.PanicError)
	}
}

func TestScanMetricView_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanMetricView(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanMetricFromRows tests (write side — takes pgx.Rows)
// ---------------------------------------------------------------------------

func TestScanMetricFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillMetricDest(dest, id, now)
		return nil
	}}

	fm, err := scanMetricFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.ID() != id {
		t.Errorf("ID = %v, want %v", fm.ID(), id)
	}
	if fm.Name() != "handleRequest" {
		t.Errorf("Name = %q, want %q", fm.Name(), "handleRequest")
	}
	if fm.LatencyMs() != 42.5 {
		t.Errorf("LatencyMs = %f, want 42.5", fm.LatencyMs())
	}
}

func TestScanMetricFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanMetricFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithName(t *testing.T) {
	name := "handleRequest"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{Name: &name})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithIsPanic(t *testing.T) {
	isPanic := true
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{IsPanic: &isPanic})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithFromDate(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour)
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{FromDate: &from})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithToDate(t *testing.T) {
	to := time.Now()
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{ToDate: &to})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	name := "fn"
	isPanic := false
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	conds := squirrel.And{}
	result := applyFilters(conds, domain.MetricFilter{Name: &name, IsPanic: &isPanic, FromDate: &from, ToDate: &to})
	if len(result) != 4 {
		t.Errorf("expected 4 conditions, got %d", len(result))
	}
}
