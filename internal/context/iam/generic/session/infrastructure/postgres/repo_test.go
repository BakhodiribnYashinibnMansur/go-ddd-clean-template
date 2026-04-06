package postgres

import (
	"testing"
	"time"

	"gct/internal/context/iam/generic/session/application/dto"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewSessionReadRepo(t *testing.T) {
	repo := NewSessionReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewSessionReadRepo_PoolIsNil(t *testing.T) {
	repo := NewSessionReadRepo(nil)
	if repo.pool != nil {
		t.Fatal("expected nil pool")
	}
}

// ---------------------------------------------------------------------------
// SessionsFilter tests (inline filter logic in List)
// ---------------------------------------------------------------------------

func TestSessionsFilter_NoFilters(t *testing.T) {
	f := dto.SessionsFilter{}
	if f.UserID != nil {
		t.Error("expected nil UserID")
	}
	if f.Revoked != nil {
		t.Error("expected nil Revoked")
	}
}

func TestSessionsFilter_UserIDOnly(t *testing.T) {
	uid := uuid.New()
	f := dto.SessionsFilter{UserID: &uid, Limit: 10}
	if f.UserID == nil || *f.UserID != uid {
		t.Errorf("expected user_id %v", uid)
	}
	if f.Revoked != nil {
		t.Error("expected nil Revoked")
	}
}

func TestSessionsFilter_RevokedOnly(t *testing.T) {
	rev := false
	f := dto.SessionsFilter{Revoked: &rev}
	if f.Revoked == nil || *f.Revoked != false {
		t.Error("expected revoked false")
	}
}

func TestSessionsFilter_AllFilters(t *testing.T) {
	uid := uuid.New()
	rev := true
	f := dto.SessionsFilter{UserID: &uid, Revoked: &rev, Limit: 20, Offset: 5}
	if *f.UserID != uid {
		t.Error("wrong user_id")
	}
	if !*f.Revoked {
		t.Error("expected revoked true")
	}
	if f.Limit != 20 {
		t.Errorf("expected limit 20, got %d", f.Limit)
	}
}

// ---------------------------------------------------------------------------
// SessionView construction test
// ---------------------------------------------------------------------------

func TestSessionView_Fields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	sid := uuid.New()
	uid := uuid.New()

	v := dto.SessionView{
		ID:           sid,
		UserID:       uid,
		DeviceID:     "device-1",
		DeviceName:   "Chrome",
		DeviceType:   "browser",
		IPAddress:    "192.168.1.100",
		UserAgent:    "Mozilla/5.0",
		ExpiresAt:    now.Add(24 * time.Hour),
		LastActivity: now,
		Revoked:      false,
		CreatedAt:    now,
	}

	if v.ID != sid {
		t.Errorf("expected ID %v, got %v", sid, v.ID)
	}
	if v.DeviceID != "device-1" {
		t.Errorf("expected device_id 'device-1', got %q", v.DeviceID)
	}
	if v.IPAddress != "192.168.1.100" {
		t.Errorf("expected ip '192.168.1.100', got %q", v.IPAddress)
	}
	if v.Revoked {
		t.Error("expected revoked false")
	}
}
