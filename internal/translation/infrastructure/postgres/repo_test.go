package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/translation/domain"

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
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewTranslationWriteRepo(t *testing.T) {
	repo := NewTranslationWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewTranslationReadRepo(t *testing.T) {
	repo := NewTranslationReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanTranslationView tests (read side — scans into struct fields directly)
// ---------------------------------------------------------------------------

func TestScanTranslationView_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	// read_repo scanTranslationView scans: &v.ID, &v.Key, &v.Language, &v.Value, &v.Group, &v.CreatedAt, &v.UpdatedAt
	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = id
		*dest[1].(*string) = "greeting"
		*dest[2].(*string) = "en"
		*dest[3].(*string) = "Hello"
		*dest[4].(*string) = "common"
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		return nil
	}}

	v, err := scanTranslationView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
	if v.Key != "greeting" {
		t.Errorf("Key = %q, want %q", v.Key, "greeting")
	}
	if v.Language != "en" {
		t.Errorf("Language = %q, want %q", v.Language, "en")
	}
	if v.Value != "Hello" {
		t.Errorf("Value = %q, want %q", v.Value, "Hello")
	}
	if v.Group != "common" {
		t.Errorf("Group = %q, want %q", v.Group, "common")
	}
}

func TestScanTranslationView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanTranslationView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanTranslationViewFromRows tests
// ---------------------------------------------------------------------------

func TestScanTranslationViewFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = id
		*dest[1].(*string) = "k"
		*dest[2].(*string) = "uz"
		*dest[3].(*string) = "v"
		*dest[4].(*string) = "g"
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		return nil
	}}

	v, err := scanTranslationViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != id {
		t.Errorf("ID = %v, want %v", v.ID, id)
	}
}

func TestScanTranslationViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanTranslationViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanTranslation tests (write side — scans into local vars)
// Scan order: id, entityType, entityID(uuid), langCode, data([]byte), createdAt, updatedAt
// ---------------------------------------------------------------------------

func TestScanTranslation_Success(t *testing.T) {
	id := uuid.New()
	entityID := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = id
		*dest[1].(*string) = "entity_key"
		*dest[2].(*uuid.UUID) = entityID
		*dest[3].(*string) = "en"
		*dest[4].(*[]byte) = []byte("translation value")
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		return nil
	}}

	tr, err := scanTranslation(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.ID() != id {
		t.Errorf("ID = %v, want %v", tr.ID(), id)
	}
	if tr.Key() != "entity_key" {
		t.Errorf("Key = %q, want %q", tr.Key(), "entity_key")
	}
	if tr.Language() != "en" {
		t.Errorf("Language = %q, want %q", tr.Language(), "en")
	}
	if tr.Value() != "translation value" {
		t.Errorf("Value = %q, want %q", tr.Value(), "translation value")
	}
	if tr.Group() != entityID.String() {
		t.Errorf("Group = %q, want %q", tr.Group(), entityID.String())
	}
}

func TestScanTranslation_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error { return errors.New("scan error") }}
	_, err := scanTranslation(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanTranslationFromRows tests (write side)
// ---------------------------------------------------------------------------

func TestScanTranslationFromRows_Success(t *testing.T) {
	id := uuid.New()
	entityID := uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = id
		*dest[1].(*string) = "k"
		*dest[2].(*uuid.UUID) = entityID
		*dest[3].(*string) = "ru"
		*dest[4].(*[]byte) = []byte("val")
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		return nil
	}}

	tr, err := scanTranslationFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.ID() != id {
		t.Errorf("ID = %v, want %v", tr.ID(), id)
	}
}

func TestScanTranslationFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error { return errors.New("rows scan error") }}
	_, err := scanTranslationFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// applyFilters tests
// ---------------------------------------------------------------------------

func TestApplyFilters_NoFilters(t *testing.T) {
	conds := squirrel.And{}
	result := applyFilters(conds, domain.TranslationFilter{})
	if len(result) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(result))
	}
}

func TestApplyFilters_WithKey(t *testing.T) {
	key := "greeting"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.TranslationFilter{Key: &key})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithLanguage(t *testing.T) {
	lang := "en"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.TranslationFilter{Language: &lang})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_WithGroup(t *testing.T) {
	group := "auth"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.TranslationFilter{Group: &group})
	if len(result) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result))
	}
}

func TestApplyFilters_AllFilters(t *testing.T) {
	key := "k"
	lang := "en"
	group := "g"
	conds := squirrel.And{}
	result := applyFilters(conds, domain.TranslationFilter{Key: &key, Language: &lang, Group: &group})
	if len(result) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(result))
	}
}
