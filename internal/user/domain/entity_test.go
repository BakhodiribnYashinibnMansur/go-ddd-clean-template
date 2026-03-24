package domain_test

import (
	"errors"
	"testing"

	domain "gct/internal/user/domain"

	"github.com/google/uuid"
)

func mustPhone(t *testing.T, raw string) domain.Phone {
	t.Helper()
	p, err := domain.NewPhone(raw)
	if err != nil {
		t.Fatalf("NewPhone(%q): %v", raw, err)
	}
	return p
}

func mustPassword(t *testing.T, raw string) domain.Password {
	t.Helper()
	pw, err := domain.NewPasswordFromRaw(raw)
	if err != nil {
		t.Fatalf("NewPasswordFromRaw: %v", err)
	}
	return pw
}

func TestNewUser_Defaults(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	if u.Phone().Value() != "+998901234567" {
		t.Fatal("phone mismatch")
	}
	if !u.IsActive() {
		t.Fatal("new user should be active")
	}
	if u.IsApproved() {
		t.Fatal("new user should not be approved")
	}
	if u.Email() != nil {
		t.Fatal("email should be nil by default")
	}
	if u.Username() != nil {
		t.Fatal("username should be nil by default")
	}
	if u.RoleID() != nil {
		t.Fatal("roleID should be nil by default")
	}
	if len(u.Sessions()) != 0 {
		t.Fatal("sessions should be empty")
	}
	if u.Attributes() == nil {
		t.Fatal("attributes should be initialized (not nil)")
	}

	// Should have a UserCreated event
	events := u.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "user.created" {
		t.Fatalf("expected user.created, got %s", events[0].EventName())
	}
}

func TestNewUser_WithOptions(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	email, _ := domain.NewEmail("test@example.com")
	roleID := uuid.New()

	u := domain.NewUser(phone, pw,
		domain.WithEmail(email),
		domain.WithUsername("john"),
		domain.WithRoleID(roleID),
		domain.WithAttributes(map[string]any{"level": 5}),
	)

	if u.Email() == nil || u.Email().Value() != "test@example.com" {
		t.Fatal("email option not applied")
	}
	if u.Username() == nil || *u.Username() != "john" {
		t.Fatal("username option not applied")
	}
	if u.RoleID() == nil || *u.RoleID() != roleID {
		t.Fatal("roleID option not applied")
	}
	if u.Attributes()["level"] != 5 {
		t.Fatal("attributes option not applied")
	}
}

func TestUser_AddSession(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents() // clear the UserCreated event

	s, err := u.AddSession(domain.DeviceDesktop, "1.2.3.4", "TestAgent")
	if err != nil {
		t.Fatalf("AddSession: %v", err)
	}
	if s == nil {
		t.Fatal("session should not be nil")
	}
	if len(u.Sessions()) != 1 {
		t.Fatalf("expected 1 session, got %d", len(u.Sessions()))
	}

	// Should have a UserSignedIn event
	events := u.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "user.signed_in" {
		t.Fatalf("expected user.signed_in, got %s", events[0].EventName())
	}
}

func TestUser_AddSession_MaxReached(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	for i := 0; i < 10; i++ {
		_, err := u.AddSession(domain.DeviceMobile, "1.1.1.1", "Agent")
		if err != nil {
			t.Fatalf("AddSession %d: %v", i, err)
		}
	}
	_, err := u.AddSession(domain.DeviceMobile, "1.1.1.1", "Agent")
	if !errors.Is(err, domain.ErrMaxSessionsReached) {
		t.Fatalf("expected ErrMaxSessionsReached, got %v", err)
	}
}

func TestUser_RemoveSession(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	s, _ := u.AddSession(domain.DeviceDesktop, "1.1.1.1", "Agent")
	if err := u.RemoveSession(s.ID()); err != nil {
		t.Fatalf("RemoveSession: %v", err)
	}
	if len(u.Sessions()) != 0 {
		t.Fatal("sessions should be empty after removal")
	}
}

func TestUser_RemoveSession_NotFound(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	err := u.RemoveSession(uuid.New())
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestUser_RevokeAllSessions(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.AddSession(domain.DeviceDesktop, "1.1.1.1", "A1")
	u.AddSession(domain.DeviceMobile, "2.2.2.2", "A2")

	u.RevokeAllSessions()
	for _, s := range u.Sessions() {
		if !s.IsRevoked() {
			t.Fatal("all sessions should be revoked")
		}
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	if err := u.VerifyPassword("SecureP@ss1"); err != nil {
		t.Fatalf("VerifyPassword should succeed: %v", err)
	}
	if err := u.VerifyPassword("WrongPassword"); err == nil {
		t.Fatal("VerifyPassword should fail for wrong password")
	}
}

func TestUser_ChangePassword(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "OldPassword1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents()

	err := u.ChangePassword("OldPassword1", "NewPassword1")
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if err := u.VerifyPassword("NewPassword1"); err != nil {
		t.Fatal("new password should work after change")
	}
	if err := u.VerifyPassword("OldPassword1"); err == nil {
		t.Fatal("old password should no longer work")
	}

	events := u.Events()
	if len(events) != 1 || events[0].EventName() != "user.password_changed" {
		t.Fatal("expected password_changed event")
	}
}

func TestUser_ChangePassword_WrongOld(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "CorrectOld1")
	u := domain.NewUser(phone, pw)

	err := u.ChangePassword("WrongOld123", "NewPassword1")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestUser_Activate_Deactivate(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents()

	u.Deactivate()
	if u.IsActive() {
		t.Fatal("user should be inactive")
	}
	events := u.Events()
	if len(events) != 1 || events[0].EventName() != "user.deactivated" {
		t.Fatal("expected user.deactivated event")
	}

	u.Activate()
	if !u.IsActive() {
		t.Fatal("user should be active after Activate()")
	}
}

func TestUser_Approve(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents()

	u.Approve()
	if !u.IsApproved() {
		t.Fatal("user should be approved")
	}
	events := u.Events()
	if len(events) != 1 || events[0].EventName() != "user.approved" {
		t.Fatal("expected user.approved event")
	}
}

func TestUser_ChangeRole(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents()

	roleID := uuid.New()
	u.ChangeRole(roleID)
	if u.RoleID() == nil || *u.RoleID() != roleID {
		t.Fatal("role ID should be updated")
	}
	events := u.Events()
	if len(events) != 1 || events[0].EventName() != "user.role_changed" {
		t.Fatal("expected user.role_changed event")
	}
}

func TestUser_UpdateLastSeen(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	if u.LastSeen() != nil {
		t.Fatal("lastSeen should be nil initially")
	}
	u.UpdateLastSeen()
	if u.LastSeen() == nil {
		t.Fatal("lastSeen should be set after UpdateLastSeen()")
	}
}
