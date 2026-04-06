package postgres

import (
	"errors"
	"testing"
	"time"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"

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

func (m *mockRows) Scan(dest ...any) error                       { return m.scanFunc(dest...) }
func (m *mockRows) Close()                                       {}
func (m *mockRows) Err() error                                   { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool                                   { return false }
func (m *mockRows) Values() ([]any, error)                       { return nil, nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }

// populateIntegrationScan fills the base 7 columns plus the 11 JWT columns
// with zero/NULL values. Individual tests override indices 0..6 with their
// own values before delegating to this helper.
func populateJWTScanDefaults(dest []any) {
	// jwt_api_key_hash (BYTEA)
	*dest[7].(*[]byte) = nil
	// jwt_access_ttl_seconds (*int)
	*dest[8].(**int) = nil
	// jwt_refresh_ttl_seconds (*int)
	*dest[9].(**int) = nil
	// jwt_public_key_pem (*string)
	*dest[10].(**string) = nil
	// jwt_previous_public_key_pem (*string)
	*dest[11].(**string) = nil
	// jwt_key_id (*string)
	*dest[12].(**string) = nil
	// jwt_previous_key_id (*string)
	*dest[13].(**string) = nil
	// jwt_rotated_at (*time.Time)
	*dest[14].(**time.Time) = nil
	// jwt_rotate_every_days (int)
	*dest[15].(*int) = 30
	// jwt_binding_mode (string)
	*dest[16].(*string) = "warn"
	// jwt_max_sessions (int)
	*dest[17].(*int) = 0
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewIntegrationWriteRepo(t *testing.T) {
	repo := NewIntegrationWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewIntegrationReadRepo(t *testing.T) {
	repo := NewIntegrationReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

// ---------------------------------------------------------------------------
// scanIntegration (write_repo.go — pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanIntegration_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)
	desc := "Slack integration"

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "Slack"
		*dest[2].(**string) = &desc
		*dest[3].(*string) = "https://hooks.slack.com/test"
		*dest[4].(*bool) = true
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		populateJWTScanDefaults(dest)
		return nil
	}}

	i, err := scanIntegration(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i == nil {
		t.Fatal("expected non-nil Integration")
	}
	if i.ID() != testID {
		t.Errorf("expected ID %v, got %v", testID, i.ID())
	}
	if i.Name() != "Slack" {
		t.Errorf("expected name 'Slack', got %q", i.Name())
	}
	if i.WebhookURL() != "https://hooks.slack.com/test" {
		t.Errorf("expected webhookURL, got %q", i.WebhookURL())
	}
	if !i.Enabled() {
		t.Error("expected enabled true")
	}
	if i.JWTBindingMode() != "warn" {
		t.Errorf("expected binding mode 'warn', got %q", i.JWTBindingMode())
	}
	if i.JWTRotateEveryDays() != 30 {
		t.Errorf("expected rotate every days 30, got %d", i.JWTRotateEveryDays())
	}
}

func TestScanIntegration_NilDescription(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "Webhook"
		*dest[2].(**string) = nil
		*dest[3].(*string) = "https://example.com/webhook"
		*dest[4].(*bool) = false
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		populateJWTScanDefaults(dest)
		return nil
	}}

	i, err := scanIntegration(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i == nil {
		t.Fatal("expected non-nil Integration")
	}
}

func TestScanIntegration_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("scan error")
	}}

	_, err := scanIntegration(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanIntegrationView (read_repo.go — pgx.Row) — success
// ---------------------------------------------------------------------------

func TestScanIntegrationView_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)
	desc := "GitHub integration"

	row := &mockRow{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "GitHub"
		*dest[2].(**string) = &desc
		*dest[3].(*string) = "https://api.github.com"
		*dest[4].(*bool) = true
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		populateJWTScanDefaults(dest)
		return nil
	}}

	v, err := scanIntegrationView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != integentity.IntegrationID(testID) {
		t.Errorf("expected ID %v, got %v", testID, v.ID)
	}
	if v.Name != "GitHub" {
		t.Errorf("expected name 'GitHub', got %q", v.Name)
	}
	if v.Type != "GitHub integration" {
		t.Errorf("expected type 'GitHub integration', got %q", v.Type)
	}
	if v.WebhookURL != "https://api.github.com" {
		t.Errorf("expected webhook URL, got %q", v.WebhookURL)
	}
	if !v.Enabled {
		t.Error("expected enabled true")
	}
	if v.HasJWT {
		t.Error("expected HasJWT false when hash is nil")
	}
}

func TestScanIntegrationView_Error(t *testing.T) {
	row := &mockRow{scanFunc: func(dest ...any) error {
		return errors.New("view scan error")
	}}

	_, err := scanIntegrationView(row)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanIntegrationViewFromRows (pgx.Rows) — success
// ---------------------------------------------------------------------------

func TestScanIntegrationViewFromRows_Success(t *testing.T) {
	testID := uuid.New()
	now := time.Now().Truncate(time.Second)

	rows := &mockRows{scanFunc: func(dest ...any) error {
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "PagerDuty"
		*dest[2].(**string) = nil
		*dest[3].(*string) = "https://events.pagerduty.com"
		*dest[4].(*bool) = false
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		populateJWTScanDefaults(dest)
		return nil
	}}

	v, err := scanIntegrationViewFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "PagerDuty" {
		t.Errorf("expected name 'PagerDuty', got %q", v.Name)
	}
	if v.Type != "" {
		t.Errorf("expected empty type for nil description, got %q", v.Type)
	}
}

func TestScanIntegrationViewFromRows_Error(t *testing.T) {
	rows := &mockRows{scanFunc: func(dest ...any) error {
		return errors.New("rows scan error")
	}}

	_, err := scanIntegrationViewFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}
