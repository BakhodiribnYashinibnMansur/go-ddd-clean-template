package postgres

import (
	"errors"
	"gct/internal/context/iam/generic/authz/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ---------------------------------------------------------------------------
// Mock helpers
// ---------------------------------------------------------------------------

type authzMockRow struct {
	scanFunc func(dest ...any) error
}

func (m *authzMockRow) Scan(dest ...any) error { return m.scanFunc(dest...) }

type authzMockRows struct {
	scanFunc func(dest ...any) error
}

func (m *authzMockRows) Scan(dest ...any) error                       { return m.scanFunc(dest...) }
func (m *authzMockRows) Close()                                       {}
func (m *authzMockRows) Err() error                                   { return nil }
func (m *authzMockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *authzMockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *authzMockRows) Next() bool                                   { return false }
func (m *authzMockRows) Values() ([]any, error)                       { return nil, nil }
func (m *authzMockRows) RawValues() [][]byte                          { return nil }
func (m *authzMockRows) Conn() *pgx.Conn                              { return nil }

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewRoleWriteRepo(t *testing.T) {
	repo := NewRoleWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil RoleWriteRepo")
	}
}

func TestNewPermissionWriteRepo(t *testing.T) {
	repo := NewPermissionWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil PermissionWriteRepo")
	}
}

func TestNewPolicyWriteRepo(t *testing.T) {
	repo := NewPolicyWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil PolicyWriteRepo")
	}
}

func TestNewScopeWriteRepo(t *testing.T) {
	repo := NewScopeWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil ScopeWriteRepo")
	}
}

func TestNewRolePermissionRepo(t *testing.T) {
	repo := NewRolePermissionRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil RolePermissionRepo")
	}
}

func TestNewPermissionScopeRepo(t *testing.T) {
	repo := NewPermissionScopeRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil PermissionScopeRepo")
	}
}

func TestNewAuthzReadRepo(t *testing.T) {
	repo := NewAuthzReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil AuthzReadRepo")
	}
}

// ---------------------------------------------------------------------------
// scanRole tests
// ---------------------------------------------------------------------------

func TestScanRole_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	desc := "admin role"

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			// columns: id, name, description, created_at, updated_at
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*string) = "admin"
			*dest[2].(**string) = &desc
			*dest[3].(*interface{}) = now
			*dest[4].(*interface{}) = now
			return nil
		},
	}

	role, err := scanRole(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role == nil {
		t.Fatal("expected non-nil role")
	}
	if role.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, role.ID())
	}
	if role.Name() != "admin" {
		t.Fatalf("expected name 'admin', got %q", role.Name())
	}
	if role.Description() == nil || *role.Description() != desc {
		t.Fatalf("expected description %q, got %v", desc, role.Description())
	}
}

func TestScanRole_Error(t *testing.T) {
	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanRole(row)
	if err == nil {
		t.Fatal("expected error from scanRole")
	}
}

func TestScanRole_NilDescription(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*string) = "viewer"
			*dest[2].(**string) = nil
			*dest[3].(*interface{}) = now
			*dest[4].(*interface{}) = now
			return nil
		},
	}

	role, err := scanRole(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Description() != nil {
		t.Fatalf("expected nil description, got %v", role.Description())
	}
}

// ---------------------------------------------------------------------------
// scanRoleFromRows tests
// ---------------------------------------------------------------------------

func TestScanRoleFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*string) = "editor"
			*dest[2].(**string) = nil
			*dest[3].(*interface{}) = now
			*dest[4].(*interface{}) = now
			return nil
		},
	}

	role, err := scanRoleFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name() != "editor" {
		t.Fatalf("expected name 'editor', got %q", role.Name())
	}
}

func TestScanRoleFromRows_Error(t *testing.T) {
	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan failed")
		},
	}

	_, err := scanRoleFromRows(rows)
	if err == nil {
		t.Fatal("expected error from scanRoleFromRows")
	}
}

// ---------------------------------------------------------------------------
// scanPermission tests
// ---------------------------------------------------------------------------

func TestScanPermission_Success(t *testing.T) {
	id := uuid.New()
	parentID := uuid.New()
	now := time.Now()
	desc := "read users"

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			// columns: id, parent_id, name, description, created_at, updated_at
			*dest[0].(*uuid.UUID) = id
			*dest[1].(**uuid.UUID) = &parentID
			*dest[2].(*string) = "users:read"
			*dest[3].(**string) = &desc
			*dest[4].(*interface{}) = now
			*dest[5].(*interface{}) = now
			return nil
		},
	}

	perm, err := scanPermission(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, perm.ID())
	}
	if perm.ParentID() == nil || *perm.ParentID() != parentID {
		t.Fatalf("expected parentID %v, got %v", parentID, perm.ParentID())
	}
	if perm.Name() != "users:read" {
		t.Fatalf("expected name 'users:read', got %q", perm.Name())
	}
}

func TestScanPermission_NilParentAndDescription(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(**uuid.UUID) = nil
			*dest[2].(*string) = "root"
			*dest[3].(**string) = nil
			*dest[4].(*interface{}) = now
			*dest[5].(*interface{}) = now
			return nil
		},
	}

	perm, err := scanPermission(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.ParentID() != nil {
		t.Fatalf("expected nil parentID")
	}
	if perm.Description() != nil {
		t.Fatalf("expected nil description")
	}
}

func TestScanPermission_Error(t *testing.T) {
	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanPermission(row)
	if err == nil {
		t.Fatal("expected error from scanPermission")
	}
}

// ---------------------------------------------------------------------------
// scanPermissionFromRows tests
// ---------------------------------------------------------------------------

func TestScanPermissionFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(**uuid.UUID) = nil
			*dest[2].(*string) = "perm1"
			*dest[3].(**string) = nil
			*dest[4].(*interface{}) = now
			*dest[5].(*interface{}) = now
			return nil
		},
	}

	perm, err := scanPermissionFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.Name() != "perm1" {
		t.Fatalf("expected name 'perm1', got %q", perm.Name())
	}
}

func TestScanPermissionFromRows_Error(t *testing.T) {
	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan failed")
		},
	}

	_, err := scanPermissionFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// scanPolicy tests
// ---------------------------------------------------------------------------

func TestScanPolicy_Success(t *testing.T) {
	id := uuid.New()
	permID := domain.NewPermissionID()
	now := time.Now()

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			// columns: id, permission_id, effect, priority, active, created_at, updated_at
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*uuid.UUID) = permID.UUID()
			*dest[2].(*string) = "ALLOW"
			*dest[3].(*int) = 10
			*dest[4].(*bool) = true
			*dest[5].(*interface{}) = now
			*dest[6].(*interface{}) = now
			return nil
		},
	}

	policy, err := scanPolicy(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, policy.ID())
	}
	if policy.PermissionID() != permID.UUID() {
		t.Fatalf("expected PermissionID %v, got %v", permID, policy.PermissionID())
	}
	if policy.Effect() != "ALLOW" {
		t.Fatalf("expected effect ALLOW, got %v", policy.Effect())
	}
	if policy.Priority() != 10 {
		t.Fatalf("expected priority 10, got %d", policy.Priority())
	}
	if !policy.IsActive() {
		t.Fatal("expected active=true")
	}
	// Conditions are loaded via metadata repo, not scanned; scanPolicy returns nil conditions.
	if len(policy.Conditions()) != 0 {
		t.Fatalf("expected empty conditions from scan, got %v", policy.Conditions())
	}
}

func TestScanPolicy_EmptyConditions(t *testing.T) {
	id := uuid.New()
	permID := domain.NewPermissionID()
	now := time.Now()

	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*uuid.UUID) = permID.UUID()
			*dest[2].(*string) = "DENY"
			*dest[3].(*int) = 0
			*dest[4].(*bool) = false
			*dest[5].(*interface{}) = now
			*dest[6].(*interface{}) = now
			return nil
		},
	}

	policy, err := scanPolicy(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Effect() != "DENY" {
		t.Fatalf("expected effect DENY, got %v", policy.Effect())
	}
	if len(policy.Conditions()) != 0 {
		t.Fatalf("expected empty conditions, got %v", policy.Conditions())
	}
}

func TestScanPolicy_Error(t *testing.T) {
	row := &authzMockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanPolicy(row)
	if err == nil {
		t.Fatal("expected error from scanPolicy")
	}
}

// ---------------------------------------------------------------------------
// scanPolicyFromRows tests
// ---------------------------------------------------------------------------

func TestScanPolicyFromRows_Success(t *testing.T) {
	id := uuid.New()
	permID := domain.NewPermissionID()
	now := time.Now()

	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*uuid.UUID) = permID.UUID()
			*dest[2].(*string) = "ALLOW"
			*dest[3].(*int) = 5
			*dest[4].(*bool) = true
			*dest[5].(*interface{}) = now
			*dest[6].(*interface{}) = now
			return nil
		},
	}

	policy, err := scanPolicyFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Priority() != 5 {
		t.Fatalf("expected priority 5, got %d", policy.Priority())
	}
}

func TestScanPolicyFromRows_Error(t *testing.T) {
	rows := &authzMockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan failed")
		},
	}

	_, err := scanPolicyFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// toTime tests
// ---------------------------------------------------------------------------

func TestToTime_WithTimeValue(t *testing.T) {
	now := time.Now()
	result := toTime(now)
	if !result.Equal(now) {
		t.Fatalf("expected %v, got %v", now, result)
	}
}

func TestToTime_WithNil(t *testing.T) {
	result := toTime(nil)
	if !result.IsZero() {
		t.Fatalf("expected zero time for nil, got %v", result)
	}
}

func TestToTime_WithNonTimeValue(t *testing.T) {
	result := toTime("not a time")
	if !result.IsZero() {
		t.Fatalf("expected zero time for non-time value, got %v", result)
	}
}
