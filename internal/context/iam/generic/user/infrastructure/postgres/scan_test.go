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
// Mock helpers — prefixed to avoid collision with write_repo_test.go
// ---------------------------------------------------------------------------

type userScanMockRow struct {
	scanFunc func(dest ...any) error
}

func (m *userScanMockRow) Scan(dest ...any) error { return m.scanFunc(dest...) }

type userScanMockRows struct {
	scanFunc func(dest ...any) error
}

func (m *userScanMockRows) Scan(dest ...any) error                       { return m.scanFunc(dest...) }
func (m *userScanMockRows) Close()                                       {}
func (m *userScanMockRows) Err() error                                   { return nil }
func (m *userScanMockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *userScanMockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *userScanMockRows) Next() bool                                   { return false }
func (m *userScanMockRows) Values() ([]any, error)                       { return nil, nil }
func (m *userScanMockRows) RawValues() [][]byte                          { return nil }
func (m *userScanMockRows) Conn() *pgx.Conn                             { return nil }

// ---------------------------------------------------------------------------
// scanUser tests
// ---------------------------------------------------------------------------

func populateUserDest(dest []any, id uuid.UUID, roleID *uuid.UUID, username *string, email *string, phone, pwHash string, active, isApproved bool, createdAt, updatedAt time.Time, deletedAt int64, lastSeen *time.Time) {
	// columns: id, role_id, username, email, phone, password_hash, salt, active, is_approved, created_at, updated_at, deleted_at, last_seen
	*dest[0].(*uuid.UUID) = id
	*dest[1].(**uuid.UUID) = roleID
	*dest[2].(**string) = username
	*dest[3].(**string) = email
	*dest[4].(*string) = phone
	*dest[5].(*string) = pwHash
	*dest[6].(**string) = nil // salt
	*dest[7].(*bool) = active
	*dest[8].(*bool) = isApproved
	*dest[9].(*time.Time) = createdAt
	*dest[10].(*time.Time) = updatedAt
	*dest[11].(*int64) = deletedAt
	*dest[12].(**time.Time) = lastSeen
}

func TestScanUser_Success(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()
	now := time.Now()
	username := "testuser"
	email := "test@example.com"
	phone := "+998901234567"

	row := &userScanMockRow{
		scanFunc: func(dest ...any) error {
			populateUserDest(dest, id, &roleID, &username, &email, phone, "$2a$10$hash", true, true, now, now, 0, nil)
			return nil
		},
	}

	user, err := scanUser(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, user.ID())
	}
	if user.RoleID() == nil || *user.RoleID() != roleID {
		t.Fatalf("expected roleID %v, got %v", roleID, user.RoleID())
	}
	if user.Username() == nil || *user.Username() != username {
		t.Fatalf("expected username %q, got %v", username, user.Username())
	}
	if !user.IsActive() {
		t.Fatal("expected active=true")
	}
	if !user.IsApproved() {
		t.Fatal("expected isApproved=true")
	}
	if user.DeletedAt() != nil {
		t.Fatalf("expected nil deletedAt for zero unix, got %v", user.DeletedAt())
	}
}

func TestScanUser_NilOptionalFields(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	row := &userScanMockRow{
		scanFunc: func(dest ...any) error {
			populateUserDest(dest, id, nil, nil, nil, "+998901234567", "$2a$10$hash", false, false, now, now, 0, nil)
			return nil
		},
	}

	user, err := scanUser(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.RoleID() != nil {
		t.Fatalf("expected nil roleID, got %v", user.RoleID())
	}
	if user.Username() != nil {
		t.Fatalf("expected nil username, got %v", user.Username())
	}
}

func TestScanUser_WithDeletedAt(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	deletedUnix := now.Unix()

	row := &userScanMockRow{
		scanFunc: func(dest ...any) error {
			populateUserDest(dest, id, nil, nil, nil, "+998901234567", "$2a$10$hash", false, false, now, now, deletedUnix, nil)
			return nil
		},
	}

	user, err := scanUser(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.DeletedAt() == nil {
		t.Fatal("expected non-nil deletedAt")
	}
}

func TestScanUser_Error(t *testing.T) {
	row := &userScanMockRow{
		scanFunc: func(dest ...any) error {
			return errors.New("scan failed")
		},
	}

	_, err := scanUser(row)
	if err == nil {
		t.Fatal("expected error from scanUser")
	}
}

// ---------------------------------------------------------------------------
// scanUserFromRows tests
// ---------------------------------------------------------------------------

func TestScanUserFromRows_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	username := "rowuser"

	rows := &userScanMockRows{
		scanFunc: func(dest ...any) error {
			populateUserDest(dest, id, nil, &username, nil, "+998901234567", "$2a$10$hash", true, false, now, now, 0, nil)
			return nil
		},
	}

	user, err := scanUserFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username() == nil || *user.Username() != "rowuser" {
		t.Fatalf("expected username 'rowuser', got %v", user.Username())
	}
}

func TestScanUserFromRows_Error(t *testing.T) {
	rows := &userScanMockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("rows scan failed")
		},
	}

	_, err := scanUserFromRows(rows)
	if err == nil {
		t.Fatal("expected error from scanUserFromRows")
	}
}

// ---------------------------------------------------------------------------
// scanSessionFromRows tests
// ---------------------------------------------------------------------------

func TestScanSessionFromRows_Success(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	deviceID := "dev-123"
	deviceName := "iPhone"
	deviceType := "mobile"
	ipAddr := "192.168.1.1"
	ua := "Mozilla/5.0"
	refreshHash := "hash123"

	rows := &userScanMockRows{
		scanFunc: func(dest ...any) error {
			// columns: id, user_id, device_id, device_name, device_type, ip_address::text, user_agent, refresh_token_hash, expires_at, last_activity, revoked, created_at, updated_at
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*uuid.UUID) = userID
			*dest[2].(**string) = &deviceID
			*dest[3].(**string) = &deviceName
			*dest[4].(**string) = &deviceType
			*dest[5].(**string) = &ipAddr
			*dest[6].(**string) = &ua
			*dest[7].(**string) = &refreshHash
			*dest[8].(*time.Time) = now.Add(24 * time.Hour)
			*dest[9].(*time.Time) = now
			*dest[10].(*bool) = false
			*dest[11].(*time.Time) = now
			*dest[12].(*time.Time) = now
			return nil
		},
	}

	session, err := scanSessionFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, session.ID())
	}
	if session.UserID() != userID {
		t.Fatalf("expected userID %v, got %v", userID, session.UserID())
	}
	if session.DeviceID() != deviceID {
		t.Fatalf("expected deviceID %q, got %q", deviceID, session.DeviceID())
	}
	if session.DeviceName() != deviceName {
		t.Fatalf("expected deviceName %q, got %q", deviceName, session.DeviceName())
	}
	if session.IPAddress().String() != ipAddr {
		t.Fatalf("expected ipAddress %q, got %q", ipAddr, session.IPAddress().String())
	}
	if session.UserAgent().String() != ua {
		t.Fatalf("expected userAgent %q, got %q", ua, session.UserAgent().String())
	}
	if session.IsRevoked() {
		t.Fatal("expected revoked=false")
	}
}

func TestScanSessionFromRows_NilOptionalFields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	rows := &userScanMockRows{
		scanFunc: func(dest ...any) error {
			*dest[0].(*uuid.UUID) = id
			*dest[1].(*uuid.UUID) = userID
			*dest[2].(**string) = nil
			*dest[3].(**string) = nil
			*dest[4].(**string) = nil
			*dest[5].(**string) = nil
			*dest[6].(**string) = nil
			*dest[7].(**string) = nil
			*dest[8].(*time.Time) = now.Add(24 * time.Hour)
			*dest[9].(*time.Time) = now
			*dest[10].(*bool) = true
			*dest[11].(*time.Time) = now
			*dest[12].(*time.Time) = now
			return nil
		},
	}

	session, err := scanSessionFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.DeviceID() != "" {
		t.Fatalf("expected empty deviceID for nil, got %q", session.DeviceID())
	}
	if session.IPAddress().String() != "" {
		t.Fatalf("expected empty ipAddress for nil, got %q", session.IPAddress().String())
	}
	if !session.IsRevoked() {
		t.Fatal("expected revoked=true")
	}
}

func TestScanSessionFromRows_Error(t *testing.T) {
	rows := &userScanMockRows{
		scanFunc: func(dest ...any) error {
			return errors.New("session scan failed")
		},
	}

	_, err := scanSessionFromRows(rows)
	if err == nil {
		t.Fatal("expected error from scanSessionFromRows")
	}
}

// ---------------------------------------------------------------------------
// reconstructUserFromRow tests
// ---------------------------------------------------------------------------

func TestReconstructUserFromRow_WithEmail(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()
	now := time.Now()
	email := "user@test.com"
	username := "tester"

	user := reconstructUserFromRow(
		id, &roleID, &username, &email, "+998901234567", "$2a$10$hash",
		true, true, now, now, 0, nil,
	)

	if user.ID() != id {
		t.Fatalf("expected ID %v, got %v", id, user.ID())
	}
	if user.Email() == nil {
		t.Fatal("expected non-nil email")
	}
}

func TestReconstructUserFromRow_NilAttributes(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	user := reconstructUserFromRow(
		id, nil, nil, nil, "+998901234567", "$2a$10$hash",
		false, false, now, now, 0, nil,
	)

	// Attributes are loaded separately via metadata repo; reconstructUserFromRow passes nil.
	if len(user.Attributes()) != 0 {
		t.Fatalf("expected empty attributes, got %v", user.Attributes())
	}
}

func TestReconstructUserFromRow_WithDeletedAt(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	deletedUnix := now.Unix()

	user := reconstructUserFromRow(
		id, nil, nil, nil, "+998901234567", "$2a$10$hash",
		false, false, now, now, deletedUnix, nil,
	)

	if user.DeletedAt() == nil {
		t.Fatal("expected non-nil deletedAt")
	}
}

func TestReconstructUserFromRow_ZeroDeletedAt(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	user := reconstructUserFromRow(
		id, nil, nil, nil, "+998901234567", "$2a$10$hash",
		false, false, now, now, 0, nil,
	)

	if user.DeletedAt() != nil {
		t.Fatalf("expected nil deletedAt for zero unix, got %v", user.DeletedAt())
	}
}

func TestReconstructUserFromRow_InvalidEmail(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	badEmail := "not-an-email"

	user := reconstructUserFromRow(
		id, nil, nil, &badEmail, "+998901234567", "$2a$10$hash",
		false, false, now, now, 0, nil,
	)

	// Invalid email should be silently ignored
	if user.Email() != nil {
		t.Fatalf("expected nil email for invalid email string, got %v", user.Email())
	}
}

func TestReconstructUserFromRow_WithLastSeen(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	lastSeen := now.Add(-1 * time.Hour)

	user := reconstructUserFromRow(
		id, nil, nil, nil, "+998901234567", "$2a$10$hash",
		true, true, now, now, 0, &lastSeen,
	)

	if user.LastSeen() == nil {
		t.Fatal("expected non-nil lastSeen")
	}
}
