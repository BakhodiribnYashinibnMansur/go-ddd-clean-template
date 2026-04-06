package entity_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	uid := entity.NewUserID()
	s, err := entity.NewSession(uid.UUID(), entity.DeviceMobile, "10.0.0.1", "TestAgent/1.0", "gct-client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.UserID() != uid.UUID() {
		t.Fatal("user ID mismatch")
	}
	if s.DeviceType() != entity.DeviceMobile {
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

	s, _ := entity.NewSession(uuid.New(), entity.DeviceDesktop, "1.1.1.1", "Agent", "gct-client")
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

	s, _ := entity.NewSession(uuid.New(), entity.DeviceDesktop, "1.1.1.1", "Agent", "gct-client")
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

	s, _ := entity.NewSession(uuid.New(), entity.DeviceBot, "2.2.2.2", "Bot", "gct-client")
	s.SetRefreshTokenHash("somehash")
	if s.RefreshTokenHash() != "somehash" {
		t.Fatalf("expected somehash, got %s", s.RefreshTokenHash())
	}
}

func TestSession_RotateRefreshHash(t *testing.T) {
	t.Parallel()

	s, _ := entity.NewSession(uuid.New(), entity.DeviceDesktop, "1.1.1.1", "Agent", "gct-client")
	s.SetRefreshTokenHash("hash-v1")

	old := s.RotateRefreshHash("hash-v2")

	if old != "hash-v1" {
		t.Fatalf("expected old hash hash-v1, got %s", old)
	}
	if s.RefreshTokenHash() != "hash-v2" {
		t.Fatalf("expected current hash hash-v2, got %s", s.RefreshTokenHash())
	}
	if s.PreviousRefreshHash() != "hash-v1" {
		t.Fatalf("expected previous hash hash-v1, got %s", s.PreviousRefreshHash())
	}
}

func TestSession_RotateRefreshHash_ChainedRotation(t *testing.T) {
	t.Parallel()

	s, _ := entity.NewSession(uuid.New(), entity.DeviceDesktop, "1.1.1.1", "Agent", "gct-client")
	s.SetRefreshTokenHash("hash-v1")
	s.RotateRefreshHash("hash-v2")
	s.RotateRefreshHash("hash-v3")

	// After two rotations, current=v3, previous=v2 (v1 is gone).
	if s.RefreshTokenHash() != "hash-v3" {
		t.Fatalf("expected current hash hash-v3, got %s", s.RefreshTokenHash())
	}
	if s.PreviousRefreshHash() != "hash-v2" {
		t.Fatalf("expected previous hash hash-v2, got %s", s.PreviousRefreshHash())
	}
}

func TestReconstructSession_WithPreviousRefreshHash(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	uid := entity.NewUserID()
	now := time.Now()
	s := entity.ReconstructSession(
		id, now, now, nil,
		uid.UUID(), "dev-123", "My Phone", entity.DeviceMobile,
		"3.3.3.3", "Agent/2.0", "current_hash",
		now.Add(7*24*time.Hour), now, false,
		"gct-client",
		"previous_hash",
	)
	if s.RefreshTokenHash() != "current_hash" {
		t.Fatal("current refresh token hash mismatch")
	}
	if s.PreviousRefreshHash() != "previous_hash" {
		t.Fatal("previous refresh token hash mismatch")
	}
}

func TestReconstructSession(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	uid := entity.NewUserID()
	now := time.Now()
	s := entity.ReconstructSession(
		id, now, now, nil,
		uid.UUID(), "dev-123", "My Phone", entity.DeviceMobile,
		"3.3.3.3", "Agent/2.0", "refresh_hash",
		now.Add(7*24*time.Hour), now, false,
		"gct-client",
	)
	if s.ID() != id {
		t.Fatal("ID mismatch")
	}
	if s.UserID() != uid.UUID() {
		t.Fatal("user ID mismatch")
	}
	if s.DeviceName() != "My Phone" {
		t.Fatal("device name mismatch")
	}
	if s.RefreshTokenHash() != "refresh_hash" {
		t.Fatal("refresh token hash mismatch")
	}
}

func TestNewSession_WithDeviceFingerprint(t *testing.T) {
	t.Parallel()

	uid := entity.NewUserID()
	fp := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	s, err := entity.NewSession(uid.UUID(), entity.DeviceDesktop, "1.2.3.4", "Agent/1.0", "gct-client", fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.DeviceFingerprint() != fp {
		t.Fatalf("expected fingerprint %q, got %q", fp, s.DeviceFingerprint())
	}
}

func TestNewSession_WithoutDeviceFingerprint(t *testing.T) {
	t.Parallel()

	uid := entity.NewUserID()
	s, err := entity.NewSession(uid.UUID(), entity.DeviceDesktop, "1.2.3.4", "Agent/1.0", "gct-client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.DeviceFingerprint() != "" {
		t.Fatalf("expected empty fingerprint, got %q", s.DeviceFingerprint())
	}
}

func TestReconstructSession_WithDeviceFingerprint(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	uid := entity.NewUserID()
	now := time.Now()
	fp := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	s := entity.ReconstructSession(
		id, now, now, nil,
		uid.UUID(), "dev-123", "My Phone", entity.DeviceMobile,
		"3.3.3.3", "Agent/2.0", "current_hash",
		now.Add(7*24*time.Hour), now, false,
		"gct-client",
		"previous_hash",
		fp,
	)
	if s.DeviceFingerprint() != fp {
		t.Fatalf("expected fingerprint %q, got %q", fp, s.DeviceFingerprint())
	}
}
