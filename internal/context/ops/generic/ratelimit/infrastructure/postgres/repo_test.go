package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/generic/ratelimit/domain"

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
// helpers — 8 columns: id, name, path_pattern, limit_count, window_seconds, is_active, created_at, updated_at
// ---------------------------------------------------------------------------

func fillRateLimitDest(dest []any, id uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "api-global"
	*dest[2].(*string) = "/api/*"
	*dest[3].(*int) = 100
	*dest[4].(*int) = 60
	*dest[5].(*bool) = true
	*dest[6].(*time.Time) = now
	*dest[7].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewRateLimitWriteRepo(t *testing.T) {
	repo := NewRateLimitWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewRateLimitReadRepo(t *testing.T) {
	repo := NewRateLimitReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanRateLimitView tests (read side — scans into struct fields directly)
// ---------------------------------------------------------------------------

func TestScanRateLimitView_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillRateLimitDest(dest, id, now)
		return nil
	}}

	v, err := scanRateLimitView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Name != "api-global" {
		t.Errorf("Name = %q, want %q", v.Name, "api-global")
	}
	if v.Rule != "/api/*" {
		t.Errorf("Rule = %q, want %q", v.Rule, "/api/*")
	}
	if v.RequestsPerWindow != 100 {
		t.Errorf("RequestsPerWindow = %d, want 100", v.RequestsPerWindow)
	}
	if v.WindowDuration != 60 {
		t.Errorf("WindowDuration = %d, want 60", v.WindowDuration)
	}
	if !v.Enabled {
		t.Error("Enabled = false, want true")
	}
}

func TestScanRateLimitView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanRateLimitView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanRateLimitViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanRateLimitViewFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillRateLimitDest(dest, id, now)
		return nil
	}}

	v, err := scanRateLimitViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanRateLimitViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanRateLimitViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanRateLimit tests (write side)
// ---------------------------------------------------------------------------

func TestScanRateLimit_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillRateLimitDest(dest, id, now)
		return nil
	}}

	rl, err := scanRateLimit(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rl.ID() != id {
		t.Errorf("ID = %v, want %v", rl.ID(), id)
	}
	if rl.Name() != "api-global" {
		t.Errorf("Name = %q, want %q", rl.Name(), "api-global")
	}
	if rl.Rule() != "/api/*" {
		t.Errorf("Rule = %q, want %q", rl.Rule(), "/api/*")
	}
	if rl.RequestsPerWindow() != 100 {
		t.Errorf("RequestsPerWindow = %d, want 100", rl.RequestsPerWindow())
	}
	if rl.WindowDuration() != 60 {
		t.Errorf("WindowDuration = %d, want 60", rl.WindowDuration())
	}
	if !rl.Enabled() {
		t.Error("Enabled = false, want true")
	}
}

func TestScanRateLimit_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanRateLimit(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanRateLimitFromRows tests (write side)
// ---------------------------------------------------------------------------

func TestScanRateLimitFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillRateLimitDest(dest, id, now)
		return nil
	}}

	rl, err := scanRateLimitFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rl.ID() != id {
		t.Errorf("ID = %v, want %v", rl.ID(), id)
	}
}

func TestScanRateLimitFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanRateLimitFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.RateLimitFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithName(t *testing.T) {
	name := "api-global"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.RateLimitFilter{Name: &name})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithEnabled(t *testing.T) {
	enabled := true
	conds := squirrel.And{}
	result := applyFilters(conds, domain.RateLimitFilter{Enabled: &enabled})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	name := "api-global"
	enabled := false
	conds := squirrel.And{}
	result := applyFilters(conds, domain.RateLimitFilter{Name: &name, Enabled: &enabled})
	if len(result) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result))
	}
}
