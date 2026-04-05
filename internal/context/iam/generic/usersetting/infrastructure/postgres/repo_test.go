package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/generic/usersetting/domain"

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
// helpers — 6 columns: id, user_id, key, value, created_at, updated_at
// ---------------------------------------------------------------------------

func fillUserSettingDest(dest []any, id, userID uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*uuid.UUID) = userID
	*dest[2].(*string) = "theme"
	*dest[3].(*string) = "dark"
	*dest[4].(*time.Time) = now
	*dest[5].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewUserSettingWriteRepo(t *testing.T) {
	repo := NewUserSettingWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewUserSettingReadRepo(t *testing.T) {
	repo := NewUserSettingReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanUserSettingView tests (read side)
// ---------------------------------------------------------------------------

func TestScanUserSettingView_Success(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillUserSettingDest(dest, id, userID, now)
		return nil
	}}

	v, err := scanUserSettingView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.UserID != userID {
		t.Errorf("UserID = %v, want %v", v.UserID, userID)
	}
	if v.Key != "theme" {
		t.Errorf("Key = %q, want %q", v.Key, "theme")
	}
	if v.Value != "dark" {
		t.Errorf("Value = %q, want %q", v.Value, "dark")
	}
	if !v.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", v.CreatedAt, now)
	}
}

func TestScanUserSettingView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanUserSettingView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanUserSettingViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanUserSettingViewFromRows_Success(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillUserSettingDest(dest, id, userID, now)
		return nil
	}}

	v, err := scanUserSettingViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanUserSettingViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanUserSettingViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanUserSetting tests (write side)
// ---------------------------------------------------------------------------

func TestScanUserSetting_Success(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillUserSettingDest(dest, id, userID, now)
		return nil
	}}

	us, err := scanUserSetting(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if us.ID() != id {
		t.Errorf("ID = %v, want %v", us.ID(), id)
	}
	if us.UserID() != userID {
		t.Errorf("UserID = %v, want %v", us.UserID(), userID)
	}
	if us.Key() != "theme" {
		t.Errorf("Key = %q, want %q", us.Key(), "theme")
	}
	if us.Value() != "dark" {
		t.Errorf("Value = %q, want %q", us.Value(), "dark")
	}
}

func TestScanUserSetting_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanUserSetting(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.UserSettingFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithUserID(t *testing.T) {
	uid := uuid.New()
	conds := squirrel.And{}
	result := applyFilters(conds, domain.UserSettingFilter{UserID: &uid})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithKey(t *testing.T) {
	key := "theme"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.UserSettingFilter{Key: &key})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	uid := uuid.New()
	key := "locale"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.UserSettingFilter{UserID: &uid, Key: &key})
	if len(result) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result))
	}
}
