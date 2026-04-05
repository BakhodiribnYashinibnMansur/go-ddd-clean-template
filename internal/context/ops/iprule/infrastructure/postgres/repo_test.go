package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/iprule/domain"

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
// helpers — 7 columns: id, ip_address, type, reason, is_active, created_at, updated_at
// ---------------------------------------------------------------------------

func fillIPRuleDest(dest []any, id uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "192.168.1.1"
	*dest[2].(*string) = "DENY"
	*dest[3].(*string) = "suspicious activity"
	*dest[4].(*bool) = true
	*dest[5].(*time.Time) = now
	*dest[6].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewIPRuleWriteRepo(t *testing.T) {
	repo := NewIPRuleWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewIPRuleReadRepo(t *testing.T) {
	repo := NewIPRuleReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanIPRuleView tests (read side)
// ---------------------------------------------------------------------------

func TestScanIPRuleView_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillIPRuleDest(dest, id, now)
		return nil
	}}

	v, err := scanIPRuleView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.IPAddress != "192.168.1.1" {
		t.Errorf("IPAddress = %q, want %q", v.IPAddress, "192.168.1.1")
	}
	if v.Action != "DENY" {
		t.Errorf("Action = %q, want %q", v.Action, "DENY")
	}
	if v.Reason != "suspicious activity" {
		t.Errorf("Reason = %q, want %q", v.Reason, "suspicious activity")
	}
	if v.ExpiresAt != nil {
		t.Errorf("ExpiresAt = %v, want nil", v.ExpiresAt)
	}
}

func TestScanIPRuleView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanIPRuleView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanIPRuleViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanIPRuleViewFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillIPRuleDest(dest, id, now)
		return nil
	}}

	v, err := scanIPRuleViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanIPRuleViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanIPRuleViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanIPRule tests (write side)
// ---------------------------------------------------------------------------

func TestScanIPRule_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillIPRuleDest(dest, id, now)
		return nil
	}}

	rule, err := scanIPRule(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID() != id {
		t.Errorf("ID = %v, want %v", rule.ID(), id)
	}
	if rule.IPAddress() != "192.168.1.1" {
		t.Errorf("IPAddress = %q, want %q", rule.IPAddress(), "192.168.1.1")
	}
	if rule.Action() != "DENY" {
		t.Errorf("Action = %q, want %q", rule.Action(), "DENY")
	}
}

func TestScanIPRule_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanIPRule(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanIPRuleFromRows tests (write side)
// ---------------------------------------------------------------------------

func TestScanIPRuleFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillIPRuleDest(dest, id, now)
		return nil
	}}

	rule, err := scanIPRuleFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID() != id {
		t.Errorf("ID = %v, want %v", rule.ID(), id)
	}
}

func TestScanIPRuleFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanIPRuleFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.IPRuleFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithIPAddress(t *testing.T) {
	ip := "10.0.0.1"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.IPRuleFilter{IPAddress: &ip})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithAction(t *testing.T) {
	action := "ALLOW"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.IPRuleFilter{Action: &action})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	ip := "10.0.0.1"
	action := "DENY"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.IPRuleFilter{IPAddress: &ip, Action: &action})
	if len(result) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result))
	}
}
