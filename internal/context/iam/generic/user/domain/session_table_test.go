package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
)

func TestNewSession_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deviceType domain.SessionDeviceType
		ip         string
		userAgent  string
	}{
		{"desktop", domain.DeviceDesktop, "192.168.1.1", "Chrome/120"},
		{"mobile", domain.DeviceMobile, "10.0.0.1", "Safari/17"},
		{"tablet", domain.DeviceTablet, "172.16.0.1", "Firefox/119"},
		{"bot", domain.DeviceBot, "8.8.8.8", "Googlebot/2.1"},
		{"tv", domain.DeviceTV, "192.168.0.100", "SmartTV/1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userID := uuid.New()
			s, _ := domain.NewSession(userID, tt.deviceType, tt.ip, tt.userAgent)

			if s.UserID() != userID {
				t.Errorf("expected userID %s, got %s", userID, s.UserID())
			}
			if s.DeviceType() != tt.deviceType {
				t.Errorf("expected deviceType %s, got %s", tt.deviceType, s.DeviceType())
			}
			if s.IPAddress().String() != tt.ip {
				t.Errorf("expected IP %s, got %s", tt.ip, s.IPAddress().String())
			}
			if s.UserAgent().String() != tt.userAgent {
				t.Errorf("expected userAgent %s, got %s", tt.userAgent, s.UserAgent().String())
			}
			if s.IsRevoked() {
				t.Error("new session should not be revoked")
			}
			if s.IsExpired() {
				t.Error("new session should not be expired")
			}
			if !s.IsActive() {
				t.Error("new session should be active")
			}
			if s.DeviceID() == "" {
				t.Error("session should have a device ID")
			}
		})
	}
}

func TestReconstructSession_TableDriven(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name    string
		revoked bool
		expired bool // If true, expiresAt is in the past
	}{
		{"active session", false, false},
		{"revoked session", true, false},
		{"expired session", false, true},
		{"revoked and expired", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expiresAt := now.Add(7 * 24 * time.Hour)
			if tt.expired {
				expiresAt = now.Add(-1 * time.Hour)
			}

			s := domain.ReconstructSession(
				uuid.New(),
				now, now, nil,
				uuid.New(),
				"device-123", "My Device",
				domain.DeviceDesktop,
				"1.1.1.1", "TestAgent", "hash123",
				expiresAt, now,
				tt.revoked,
			)

			if s.IsRevoked() != tt.revoked {
				t.Errorf("expected revoked=%v, got %v", tt.revoked, s.IsRevoked())
			}
			if s.IsExpired() != tt.expired {
				t.Errorf("expected expired=%v, got %v", tt.expired, s.IsExpired())
			}

			expectedActive := !tt.revoked && !tt.expired
			if s.IsActive() != expectedActive {
				t.Errorf("expected active=%v, got %v", expectedActive, s.IsActive())
			}
		})
	}
}

func TestSession_Lifecycle(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	s, _ := domain.NewSession(userID, domain.DeviceDesktop, "10.0.0.1", "TestAgent")

	// Initially active
	if !s.IsActive() {
		t.Fatal("new session should be active")
	}

	// Update activity
	s.UpdateActivity()
	if s.LastActivity().IsZero() {
		t.Error("last activity should be updated")
	}

	// Set refresh token hash
	s.SetRefreshTokenHash("new-hash")
	if s.RefreshTokenHash() != "new-hash" {
		t.Errorf("expected hash 'new-hash', got %s", s.RefreshTokenHash())
	}

	// Revoke
	s.Revoke()
	if !s.IsRevoked() {
		t.Error("session should be revoked")
	}
	if s.IsActive() {
		t.Error("revoked session should not be active")
	}
}
