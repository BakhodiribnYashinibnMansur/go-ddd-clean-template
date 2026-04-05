package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	// pgproto3 not needed
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
// helpers — 8 columns: id, title, body, type, target_type, is_active, created_at, updated_at
// ---------------------------------------------------------------------------

func fillNotificationDest(dest []any, id uuid.UUID, now time.Time) {
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "New Feature"
	*dest[2].(*string) = "Check it out"
	*dest[3].(*string) = "INFO"
	*dest[4].(*string) = "all"
	*dest[5].(*bool) = true
	*dest[6].(*time.Time) = now
	*dest[7].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewNotificationWriteRepo(t *testing.T) {
	repo := NewNotificationWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewNotificationReadRepo(t *testing.T) {
	repo := NewNotificationReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanNotificationView tests (read side)
// ---------------------------------------------------------------------------

func TestScanNotificationView_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillNotificationDest(dest, id, now)
		return nil
	}}

	v, err := scanNotificationView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID.UUID() != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Title != "New Feature" {
		t.Errorf("Title = %q, want %q", v.Title, "New Feature")
	}
	if v.Message != "Check it out" {
		t.Errorf("Message = %q, want %q", v.Message, "Check it out")
	}
	if v.Type != "INFO" {
		t.Errorf("Type = %q, want %q", v.Type, "INFO")
	}
	if !v.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", v.CreatedAt, now)
	}
}

func TestScanNotificationView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanNotificationView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanNotificationViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanNotificationViewFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillNotificationDest(dest, id, now)
		return nil
	}}

	v, err := scanNotificationViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID.UUID() != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanNotificationViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanNotificationViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanNotification tests (write side)
// ---------------------------------------------------------------------------

func TestScanNotification_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillNotificationDest(dest, id, now)
		return nil
	}}

	n, err := scanNotification(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID() != id {
		t.Errorf("ID = %v, want %v", n.ID(), id)
	}
	if n.Title() != "New Feature" {
		t.Errorf("Title = %q, want %q", n.Title(), "New Feature")
	}
	if n.Message() != "Check it out" {
		t.Errorf("Message = %q, want %q", n.Message(), "Check it out")
	}
	if n.Type() != "INFO" {
		t.Errorf("Type = %q, want %q", n.Type(), "INFO")
	}
}

func TestScanNotification_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanNotification(row)
	if err == nil {
		t.Fatal("expected error")
	}
}
