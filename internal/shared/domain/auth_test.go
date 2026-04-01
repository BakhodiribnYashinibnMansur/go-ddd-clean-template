package domain_test

import (
	"testing"
	"time"

	"gct/internal/shared/domain"

	"github.com/google/uuid"
)

func TestAuthSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"expired in past", time.Now().Add(-1 * time.Hour), true},
		{"expired just now", time.Now().Add(-1 * time.Millisecond), true},
		{"not expired future", time.Now().Add(1 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := domain.AuthSession{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				DeviceID:  uuid.New(),
				ExpiresAt: tt.expiresAt,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthSession_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		revoked   bool
		want      bool
	}{
		{"active session", time.Now().Add(1 * time.Hour), false, true},
		{"expired session", time.Now().Add(-1 * time.Hour), false, false},
		{"revoked session", time.Now().Add(1 * time.Hour), true, false},
		{"expired and revoked", time.Now().Add(-1 * time.Hour), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := domain.AuthSession{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				DeviceID:  uuid.New(),
				ExpiresAt: tt.expiresAt,
				Revoked:   tt.revoked,
			}
			if got := s.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthSession_Fields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	deviceID := uuid.New()
	expires := time.Now().Add(24 * time.Hour)
	lastAct := time.Now()

	s := domain.AuthSession{
		ID:               id,
		UserID:           userID,
		DeviceID:         deviceID,
		RefreshTokenHash: "hash123",
		ExpiresAt:        expires,
		Revoked:          false,
		LastActivity:     lastAct,
	}

	if s.ID != id {
		t.Error("ID mismatch")
	}
	if s.UserID != userID {
		t.Error("UserID mismatch")
	}
	if s.DeviceID != deviceID {
		t.Error("DeviceID mismatch")
	}
	if s.RefreshTokenHash != "hash123" {
		t.Error("RefreshTokenHash mismatch")
	}
	if s.ExpiresAt != expires {
		t.Error("ExpiresAt mismatch")
	}
	if s.Revoked {
		t.Error("expected Revoked to be false")
	}
	if s.LastActivity != lastAct {
		t.Error("LastActivity mismatch")
	}
}

func TestAuthUser_Fields(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()

	u := domain.AuthUser{
		ID:         id,
		RoleID:     &roleID,
		Active:     true,
		IsApproved: true,
		Attributes: map[string]string{"key": "value"},
	}

	if u.ID != id {
		t.Error("ID mismatch")
	}
	if u.RoleID == nil || *u.RoleID != roleID {
		t.Error("RoleID mismatch")
	}
	if !u.Active {
		t.Error("expected Active to be true")
	}
	if !u.IsApproved {
		t.Error("expected IsApproved to be true")
	}
	if u.Attributes["key"] != "value" {
		t.Error("Attributes mismatch")
	}
}

func TestAuthUser_NilOptionals(t *testing.T) {
	u := domain.AuthUser{
		ID: uuid.New(),
	}

	if u.RoleID != nil {
		t.Error("expected nil RoleID")
	}
	if u.Active {
		t.Error("expected Active to be false (zero value)")
	}
	if u.Attributes != nil {
		t.Error("expected nil Attributes")
	}
}
