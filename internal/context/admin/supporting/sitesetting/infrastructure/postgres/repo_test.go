package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/sitesetting/domain"

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

func TestNewSiteSettingWriteRepo(t *testing.T) {
	repo := NewSiteSettingWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewSiteSettingReadRepo(t *testing.T) {
	repo := NewSiteSettingReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanSiteSettingView tests (read side — pgx.Row)
// ---------------------------------------------------------------------------

func TestScanSiteSettingView_Success(t *testing.T) {
	id := domain.NewSiteSettingID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id.UUID()
			*dest[1].(*string) = "site_name"
			*dest[2].(*string) = "My Site"
			*dest[3].(*string) = "general"
			*dest[4].(*string) = "The site name"
			*dest[5].(*time.Time) = now
			*dest[6].(*time.Time) = now
			return nil
		},
	}

	v, err := scanSiteSettingView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Key != "site_name" {
		t.Errorf("Key = %q, want %q", v.Key, "site_name")
	}
	if v.Value != "My Site" {
		t.Errorf("Value = %q, want %q", v.Value, "My Site")
	}
	if v.Type != "general" {
		t.Errorf("Type = %q, want %q", v.Type, "general")
	}
	if v.Description != "The site name" {
		t.Errorf("Description = %q, want %q", v.Description, "The site name")
	}
	if !v.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", v.CreatedAt, now)
	}
	if !v.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", v.UpdatedAt, now)
	}
}

func TestScanSiteSettingView_Error(t *testing.T) {
	row := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan error")
		},
	}
	_, err := scanSiteSettingView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanSiteSettingViewFromRows tests (read side — pgx.Rows)
// ---------------------------------------------------------------------------

func TestScanSiteSettingViewFromRows_Success(t *testing.T) {
	id := domain.NewSiteSettingID()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id.UUID()
			*dest[1].(*string) = "key1"
			*dest[2].(*string) = "val1"
			*dest[3].(*string) = "email"
			*dest[4].(*string) = "desc"
			*dest[5].(*time.Time) = now
			*dest[6].(*time.Time) = now
			return nil
		},
	}

	v, err := scanSiteSettingViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanSiteSettingViewFromRows_Error(t *testing.T) {
	rows := &mockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan error")
		},
	}
	_, err := scanSiteSettingViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanSiteSetting tests (write side — pgx.Row)
// ---------------------------------------------------------------------------

func TestScanSiteSetting_Success(t *testing.T) {
	id := domain.NewSiteSettingID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id.UUID()
			*dest[1].(*string) = "maint_mode"
			*dest[2].(*string) = "false"
			*dest[3].(*string) = "general"
			*dest[4].(*string) = "Maintenance mode toggle"
			*dest[5].(*time.Time) = now
			*dest[6].(*time.Time) = now
			return nil
		},
	}

	s, err := scanSiteSetting(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TypedID() != id {
		t.Errorf("ID = %v, want %v", s.TypedID(), id)
	}
	if s.Key() != "maint_mode" {
		t.Errorf("Key = %q, want %q", s.Key(), "maint_mode")
	}
	if s.Value() != "false" {
		t.Errorf("Value = %q, want %q", s.Value(), "false")
	}
	if s.Type() != "general" {
		t.Errorf("Type = %q, want %q", s.Type(), "general")
	}
	if s.Description() != "Maintenance mode toggle" {
		t.Errorf("Description = %q, want %q", s.Description(), "Maintenance mode toggle")
	}
}

func TestScanSiteSetting_Error(t *testing.T) {
	row := &mockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan error")
		},
	}
	_, err := scanSiteSetting(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanSiteSettingFromRows tests (write side — pgx.Rows)
// ---------------------------------------------------------------------------

func TestScanSiteSettingFromRows_Success(t *testing.T) {
	id := domain.NewSiteSettingID()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id.UUID()
			*dest[1].(*string) = "k"
			*dest[2].(*string) = "v"
			*dest[3].(*string) = "t"
			*dest[4].(*string) = "d"
			*dest[5].(*time.Time) = now
			*dest[6].(*time.Time) = now
			return nil
		},
	}

	s, err := scanSiteSettingFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TypedID() != id {
		t.Errorf("ID = %v, want %v", s.TypedID(), id)
	}
}

func TestScanSiteSettingFromRows_Error(t *testing.T) {
	rows := &mockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan error")
		},
	}
	_, err := scanSiteSettingFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.SiteSettingFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithKey(t *testing.T) {
	key := "site_name"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.SiteSettingFilter{Key: &key})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithType(t *testing.T) {
	sType := "general"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.SiteSettingFilter{Type: &sType})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithAllFilters(t *testing.T) {
	key := "site_name"
	sType := "general"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.SiteSettingFilter{Key: &key, Type: &sType})
	if len(result) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result))
	}
}
