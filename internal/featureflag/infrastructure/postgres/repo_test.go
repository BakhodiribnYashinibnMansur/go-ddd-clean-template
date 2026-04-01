package postgres

import (
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"

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
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewFeatureFlagWriteRepo(t *testing.T) {
	repo := NewFeatureFlagWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewFeatureFlagReadRepo(t *testing.T) {
	repo := NewFeatureFlagReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewRuleGroupWriteRepo(t *testing.T) {
	repo := NewRuleGroupWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewPgEvaluator(t *testing.T) {
	eval := NewPgEvaluator(nil)
	if eval == nil {
		t.Fatal("expected non-nil evaluator")
	}
}

// ---------------------------------------------------------------------------
// scanFeatureFlag (pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanFeatureFlag_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		// selectColumns: id, key, name, flag_type, default_value, description, rollout_percentage, is_active, created_at, updated_at
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "test-key"
		*dest[2].(*string) = "Test Flag"
		*dest[3].(*string) = "boolean"
		*dest[4].(*string) = "false"
		*dest[5].(*string) = "A test flag"
		*dest[6].(*int) = 50
		*dest[7].(*bool) = true
		*dest[8].(*time.Time) = now
		*dest[9].(*time.Time) = now
		return nil
	}}

	ff, err := scanFeatureFlag(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ff == nil {
		t.Fatal("expected non-nil FeatureFlag")
	}
	if ff.ID() != testID {
		t.Errorf("expected ID %v, got %v", testID, ff.ID())
	}
	if ff.Key() != "test-key" {
		t.Errorf("expected key 'test-key', got %q", ff.Key())
	}
	if ff.Name() != "Test Flag" {
		t.Errorf("expected name 'Test Flag', got %q", ff.Name())
	}
	if ff.FlagType() != "boolean" {
		t.Errorf("expected flag_type 'boolean', got %q", ff.FlagType())
	}
	if ff.DefaultValue() != "false" {
		t.Errorf("expected default_value 'false', got %q", ff.DefaultValue())
	}
	if ff.Description() != "A test flag" {
		t.Errorf("expected description 'A test flag', got %q", ff.Description())
	}
	if ff.RolloutPercentage() != 50 {
		t.Errorf("expected rollout 50, got %d", ff.RolloutPercentage())
	}
	if !ff.IsActive() {
		t.Error("expected is_active true")
	}
}

// ---------------------------------------------------------------------------
// scanFeatureFlag (pgx.Row) — error
// ---------------------------------------------------------------------------

func TestScanFeatureFlag_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("scan error")
	}}

	_, err := scanFeatureFlag(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanFeatureFlagFromRows (pgx.Rows) — success
// ---------------------------------------------------------------------------

func TestScanFeatureFlagFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "rows-key"
		*dest[2].(*string) = "Rows Flag"
		*dest[3].(*string) = "string"
		*dest[4].(*string) = "default"
		*dest[5].(*string) = "desc"
		*dest[6].(*int) = 100
		*dest[7].(*bool) = false
		*dest[8].(*time.Time) = now
		*dest[9].(*time.Time) = now
		return nil
	}}

	ff, err := scanFeatureFlagFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ff == nil {
		t.Fatal("expected non-nil FeatureFlag")
	}
	if ff.Key() != "rows-key" {
		t.Errorf("expected key 'rows-key', got %q", ff.Key())
	}
	if ff.RolloutPercentage() != 100 {
		t.Errorf("expected rollout 100, got %d", ff.RolloutPercentage())
	}
}

func TestScanFeatureFlagFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanFeatureFlagFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// FeatureFlagFilter (from read_repo.go List — inline filter logic)
// ---------------------------------------------------------------------------

func TestFeatureFlagFilter_NoFilters(t *testing.T) {
	filter := domain.FeatureFlagFilter{}
	// No search, no enabled — should produce no extra conditions.
	if filter.Search != nil {
		t.Error("expected nil Search")
	}
	if filter.Enabled != nil {
		t.Error("expected nil Enabled")
	}
}

func TestFeatureFlagFilter_SearchOnly(t *testing.T) {
	s := "test"
	filter := domain.FeatureFlagFilter{Search: &s, Limit: 10, Offset: 0}
	if filter.Search == nil || *filter.Search != "test" {
		t.Error("expected search 'test'")
	}
	if filter.Enabled != nil {
		t.Error("expected nil Enabled")
	}
}

func TestFeatureFlagFilter_EnabledOnly(t *testing.T) {
	enabled := true
	filter := domain.FeatureFlagFilter{Enabled: &enabled, Limit: 5}
	if filter.Enabled == nil || !*filter.Enabled {
		t.Error("expected enabled true")
	}
}

func TestFeatureFlagFilter_AllFilters(t *testing.T) {
	s := "flag"
	enabled := false
	filter := domain.FeatureFlagFilter{Search: &s, Enabled: &enabled, Limit: 20, Offset: 5}
	if *filter.Search != "flag" {
		t.Error("expected search 'flag'")
	}
	if *filter.Enabled != false {
		t.Error("expected enabled false")
	}
	if filter.Limit != 20 {
		t.Errorf("expected limit 20, got %d", filter.Limit)
	}
	if filter.Offset != 5 {
		t.Errorf("expected offset 5, got %d", filter.Offset)
	}
}
