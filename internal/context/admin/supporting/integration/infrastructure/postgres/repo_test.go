package postgres

import (
	"errors"
	"testing"
	"time"

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
		// writeColumns: id, name, description, base_url, is_active, created_at, updated_at
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "Slack"
		*dest[2].(**string) = &desc
		*dest[3].(*string) = "https://hooks.slack.com/test"
		*dest[4].(*bool) = true
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
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
		// readColumns: id, name, description, base_url, is_active, created_at, updated_at
		*dest[0].(*uuid.UUID) = testID
		*dest[1].(*string) = "GitHub"
		*dest[2].(**string) = &desc
		*dest[3].(*string) = "https://api.github.com"
		*dest[4].(*bool) = true
		*dest[5].(*time.Time) = now
		*dest[6].(*time.Time) = now
		return nil
	}}

	v, err := scanIntegrationView(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != testID {
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
