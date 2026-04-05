package domain_test

import (
	"errors"
	"testing"

	domain "gct/internal/context/iam/user/domain"
	"github.com/stretchr/testify/require"
)

func activeApprovedUser(t *testing.T) *domain.User {
	t.Helper()
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := domain.NewUser(phone, pw)
	u.Approve()
	u.ClearEvents()
	return u
}

func TestSignInService_Success(t *testing.T) {
	t.Parallel()

	svc := &domain.SignInService{}
	u := activeApprovedUser(t)

	sess, err := svc.SignIn(u, "SecureP@ss1", domain.DeviceDesktop, "10.0.0.1", "TestAgent")
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

	svc := &domain.SignInService{}
	u := activeApprovedUser(t)
	u.Deactivate()
	u.ClearEvents()

	_, err := svc.SignIn(u, "SecureP@ss1", domain.DeviceDesktop, "10.0.0.1", "TestAgent")
	if !errors.Is(err, domain.ErrUserInactive) {
		t.Fatalf("expected ErrUserInactive, got %v", err)
	}
}

func TestSignInService_NotApproved(t *testing.T) {
	t.Parallel()

	svc := &domain.SignInService{}
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := domain.NewUser(phone, pw) // not approved

	_, err := svc.SignIn(u, "SecureP@ss1", domain.DeviceDesktop, "10.0.0.1", "TestAgent")
	if !errors.Is(err, domain.ErrUserNotApproved) {
		t.Fatalf("expected ErrUserNotApproved, got %v", err)
	}
}

func TestSignInService_WrongPassword(t *testing.T) {
	t.Parallel()

	svc := &domain.SignInService{}
	u := activeApprovedUser(t)

	_, err := svc.SignIn(u, "WrongPassword", domain.DeviceDesktop, "10.0.0.1", "TestAgent")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestSignInService_MaxSessions(t *testing.T) {
	t.Parallel()

	svc := &domain.SignInService{}
	u := activeApprovedUser(t)

	for i := 0; i < 50; i++ {
		_, err := svc.SignIn(u, "SecureP@ss1", domain.DeviceMobile, "1.1.1.1", "Agent")
		if err != nil {
			t.Fatalf("SignIn %d: %v", i, err)
		}
	}
	_, err := svc.SignIn(u, "SecureP@ss1", domain.DeviceMobile, "1.1.1.1", "Agent")
	if !errors.Is(err, domain.ErrMaxSessionsReached) {
		t.Fatalf("expected ErrMaxSessionsReached, got %v", err)
	}
}
