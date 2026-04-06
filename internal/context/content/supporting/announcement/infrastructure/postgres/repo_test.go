package postgres

import (
	"errors"
	"testing"
	"time"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"

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
// helpers
// ---------------------------------------------------------------------------

func fillAnnouncementDest(dest []any, id uuid.UUID, now time.Time) {
	// id, title, content, type, is_active, priority, starts_at, ends_at, created_at, updated_at
	*dest[0].(*uuid.UUID) = id
	*dest[1].(*string) = "Title"
	*dest[2].(*string) = "Content"
	*dest[3].(*string) = "info"
	*dest[4].(*bool) = true
	*dest[5].(*int) = 1
	*dest[6].(**time.Time) = &now
	*dest[7].(**time.Time) = nil
	*dest[8].(*time.Time) = now
	*dest[9].(*time.Time) = now
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewAnnouncementWriteRepo(t *testing.T) {
	repo := NewAnnouncementWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewAnnouncementReadRepo(t *testing.T) {
	repo := NewAnnouncementReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanAnnouncementView tests (read side)
// ---------------------------------------------------------------------------

func TestScanAnnouncementView_Success(t *testing.T) {
	id := announceentity.NewAnnouncementID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillAnnouncementDest(dest, id.UUID(), now)
		return nil
	}}

	v, err := scanAnnouncementView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.TitleUz != "Title" {
		t.Errorf("TitleUz = %q, want %q", v.TitleUz, "Title")
	}
	if !v.Published {
		t.Error("Published = false, want true")
	}
	if v.Priority != 1 {
		t.Errorf("Priority = %d, want 1", v.Priority)
	}
}

func TestScanAnnouncementView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanAnnouncementView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanAnnouncementViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanAnnouncementViewFromRows_Success(t *testing.T) {
	id := announceentity.NewAnnouncementID()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillAnnouncementDest(dest, id.UUID(), now)
		return nil
	}}

	v, err := scanAnnouncementViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanAnnouncementViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanAnnouncementViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanAnnouncement tests (write side)
// ---------------------------------------------------------------------------

func TestScanAnnouncement_Success(t *testing.T) {
	id := announceentity.NewAnnouncementID()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		fillAnnouncementDest(dest, id.UUID(), now)
		return nil
	}}

	a, err := scanAnnouncement(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TypedID() != id {
		t.Errorf("ID = %v, want %v", a.ID(), id)
	}
	if a.Title().Uz != "Title" {
		t.Errorf("Title.Uz = %q, want %q", a.Title().Uz, "Title")
	}
	if !a.Published() {
		t.Error("Published = false, want true")
	}
}

func TestScanAnnouncement_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanAnnouncement(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanAnnouncementFromRows tests (write side)
// ---------------------------------------------------------------------------

func TestScanAnnouncementFromRows_Success(t *testing.T) {
	id := announceentity.NewAnnouncementID()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		fillAnnouncementDest(dest, id.UUID(), now)
		return nil
	}}

	a, err := scanAnnouncementFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TypedID() != id {
		t.Errorf("ID = %v, want %v", a.ID(), id)
	}
}

func TestScanAnnouncementFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanAnnouncementFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, announcerepo.AnnouncementFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithPublished(t *testing.T) {
	pub := true
	conds := squirrel.And{}
	result := applyFilters(conds, announcerepo.AnnouncementFilter{Published: &pub})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}
