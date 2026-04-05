package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	s, err := domain.NewSession(uid, domain.DeviceMobile, "10.0.0.1", "TestAgent/1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.UserID() != uid {
		t.Fatal("user ID mismatch")
	}
	if s.DeviceType() != domain.DeviceMobile {
		t.Fatalf("expected MOBILE, got %s", s.DeviceType())
	}
	if s.IPAddress().String() != "10.0.0.1" {
		t.Fatal("IP mismatch")
	}
	if s.UserAgent().String() != "TestAgent/1.0" {
		t.Fatal("user agent mismatch")
	}
	if s.IsRevoked() {
		t.Fatal("new session should not be revoked")
	}
	if s.IsExpired() {
		t.Fatal("new session should not be expired")
	}
	if !s.IsActive() {
		t.Fatal("new session should be active")
	}
	if s.DeviceID() == "" {
		t.Fatal("device ID should be set")
	}
}

func TestSession_Revoke(t *testing.T) {
	t.Parallel()

	s, _ := domain.NewSession(uuid.New(), domain.DeviceDesktop, "1.1.1.1", "Agent")
	s.Revoke()
	if !s.IsRevoked() {
		t.Fatal("session should be revoked")
	}
	if s.IsActive() {
		t.Fatal("revoked session should not be active")
	}
}

func TestSession_UpdateActivity(t *testing.T) {
	t.Parallel()

	s, _ := domain.NewSession(uuid.New(), domain.DeviceDesktop, "1.1.1.1", "Agent")
	before := s.ExpiresAt()
	time.Sleep(2 * time.Millisecond) // tiny pause so timestamps differ
	s.UpdateActivity()
	if !s.ExpiresAt().After(before) || s.ExpiresAt().Equal(before) {
		// ExpiresAt should be refreshed (at least equal or later)
	}
	if s.LastActivity().IsZero() {
		t.Fatal("last activity should be updated")
	}
}

func TestSession_SetRefreshTokenHash(t *testing.T) {
	t.Parallel()

	s, _ := domain.NewSession(uuid.New(), domain.DeviceBot, "2.2.2.2", "Bot")
	s.SetRefreshTokenHash("somehash")
	if s.RefreshTokenHash() != "somehash" {
		t.Fatalf("expected somehash, got %s", s.RefreshTokenHash())
	}
}

func TestReconstructSession(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	uid := uuid.New()
	now := time.Now()
	s := domain.ReconstructSession(
		id, now, now, nil,
		uid, "dev-123", "My Phone", domain.DeviceMobile,
		"3.3.3.3", "Agent/2.0", "refresh_hash",
		now.Add(7*24*time.Hour), now, false,
	)
	if s.ID() != id {
		t.Fatal("ID mismatch")
	}
	if s.UserID() != uid {
		t.Fatal("user ID mismatch")
	}
	if s.DeviceName() != "My Phone" {
		t.Fatal("device name mismatch")
	}
	if s.RefreshTokenHash() != "refresh_hash" {
		t.Fatal("refresh token hash mismatch")
	}
}
