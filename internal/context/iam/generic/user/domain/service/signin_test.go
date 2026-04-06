package service_test

import (
	"errors"
	"testing"

	"gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/context/iam/generic/user/domain/service"

	"github.com/stretchr/testify/require"
)

func mustPhone(t *testing.T, raw string) entity.Phone {
	t.Helper()
	p, err := entity.NewPhone(raw)
	if err != nil {
		t.Fatalf("NewPhone(%q): %v", raw, err)
	}
	return p
}

func mustPassword(t *testing.T, raw string) entity.Password {
	t.Helper()
	pw, err := entity.NewPasswordFromRaw(raw)
	require.NoError(t, err)
	return pw
}

func activeApprovedUser(t *testing.T) *entity.User {
	t.Helper()
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := entity.NewUser(phone, pw)
	u.Approve()
	u.ClearEvents()
	return u
}

func TestSignInService_Success(t *testing.T) {
	t.Parallel()

	svc := &service.SignInService{}
	u := activeApprovedUser(t)

	sess, err := svc.SignIn(u, "SecureP@ss1", entity.DeviceDesktop, "10.0.0.1", "TestAgent", "gct-client")
	require.NoError(t, err)
	if sess == nil {
		t.Fatal("session should not be nil")
	}
	if u.LastSeen() == nil {
		t.Fatal("lastSeen should be updated")
	}
	// Should have UserSignedIn event
	found := false
	for _, e := range u.Events() {
		if e.EventName() == "user.signed_in" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected user.signed_in event")
	}
}

func TestSignInService_Inactive(t *testing.T) {
	t.Parallel()

	svc := &service.SignInService{}
	u := activeApprovedUser(t)
	u.Deactivate()
	u.ClearEvents()

	_, err := svc.SignIn(u, "SecureP@ss1", entity.DeviceDesktop, "10.0.0.1", "TestAgent", "gct-client")
	if !errors.Is(err, entity.ErrUserInactive) {
		t.Fatalf("expected ErrUserInactive, got %v", err)
	}
}

func TestSignInService_NotApproved(t *testing.T) {
	t.Parallel()

	svc := &service.SignInService{}
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := entity.NewUser(phone, pw) // not approved

	_, err := svc.SignIn(u, "SecureP@ss1", entity.DeviceDesktop, "10.0.0.1", "TestAgent", "gct-client")
	if !errors.Is(err, entity.ErrUserNotApproved) {
		t.Fatalf("expected ErrUserNotApproved, got %v", err)
	}
}

func TestSignInService_WrongPassword(t *testing.T) {
	t.Parallel()

	svc := &service.SignInService{}
	u := activeApprovedUser(t)

	_, err := svc.SignIn(u, "WrongPassword", entity.DeviceDesktop, "10.0.0.1", "TestAgent", "gct-client")
	if !errors.Is(err, entity.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestSignInService_MaxSessions(t *testing.T) {
	t.Parallel()

	svc := &service.SignInService{}
	u := activeApprovedUser(t)

	for i := 0; i < 50; i++ {
		_, err := svc.SignIn(u, "SecureP@ss1", entity.DeviceMobile, "1.1.1.1", "Agent", "gct-client")
		if err != nil {
			t.Fatalf("SignIn %d: %v", i, err)
		}
	}
	_, err := svc.SignIn(u, "SecureP@ss1", entity.DeviceMobile, "1.1.1.1", "Agent", "gct-client")
	if !errors.Is(err, entity.ErrMaxSessionsReached) {
		t.Fatalf("expected ErrMaxSessionsReached, got %v", err)
	}
}
