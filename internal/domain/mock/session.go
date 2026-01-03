package mock

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"gct/internal/domain"
)

// Session generates a fake domain.Session
func Session() *domain.Session {
	deviceType := []domain.SessionDeviceType{
		domain.DeviceTypeDesktop,
		domain.DeviceTypeMobile,
		domain.DeviceTypeTablet,
		domain.DeviceTypeBot,
		domain.DeviceTypeTV,
	}[gofakeit.IntRange(0, 4)]

	return &domain.Session{
		ID:           UUID(),
		UserID:       UUID(),
		DeviceID:     UUID(),
		DeviceName:   func() *string { s := gofakeit.Name(); return &s }(),
		DeviceType:   func() *domain.SessionDeviceType { dt := deviceType; return &dt }(),
		IPAddress:    func() *string { ip := gofakeit.IPv4Address(); return &ip }(),
		UserAgent:    func() *string { ua := gofakeit.UserAgent(); return &ua }(),
		FCMToken:     func() *string { token := gofakeit.LetterN(20); return &token }(),
		ExpiresAt:    FutureTime(),
		LastActivity: time.Now(),
		Revoked:      false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Sessions generates multiple fake domain.Session
func Sessions(count int) []*domain.Session {
	sessions := make([]*domain.Session, count)
	for i := range count {
		sessions[i] = Session()
	}
	return sessions
}

// SessionWithUserID generates a fake domain.Session with specific UserID
func SessionWithUserID(userID uuid.UUID) *domain.Session {
	session := Session()
	session.UserID = userID
	return session
}

// SessionExpired generates a fake expired domain.Session
func SessionExpired() *domain.Session {
	session := Session()
	session.ExpiresAt = PastTime()
	session.LastActivity = PastTime()
	return session
}

// SessionRevoked generates a fake revoked domain.Session
func SessionRevoked() *domain.Session {
	session := Session()
	session.Revoked = true
	return session
}

// SessionFilter generates a fake domain.SessionFilter
func SessionFilter() *domain.SessionFilter {
	id := UUID()
	return &domain.SessionFilter{
		ID: &id,
	}
}

// SessionFilterWithID generates a domain.SessionFilter with specific ID
func SessionFilterWithID(id uuid.UUID) *domain.SessionFilter {
	return &domain.SessionFilter{
		ID: &id,
	}
}

// SessionActive generates a fake active domain.Session
func SessionActive() *domain.Session {
	session := Session()
	session.ExpiresAt = FutureTime()
	session.LastActivity = time.Now()
	session.Revoked = false
	return session
}

// SessionWithDeviceType generates a fake domain.Session with specific device type
func SessionWithDeviceType(deviceType domain.SessionDeviceType) *domain.Session {
	session := Session()
	session.DeviceType = &deviceType
	return session
}

// SessionDesktop generates a fake desktop domain.Session
func SessionDesktop() *domain.Session {
	return SessionWithDeviceType(domain.DeviceTypeDesktop)
}

// SessionMobile generates a fake mobile domain.Session
func SessionMobile() *domain.Session {
	return SessionWithDeviceType(domain.DeviceTypeMobile)
}

// SessionTablet generates a fake tablet domain.Session
func SessionTablet() *domain.Session {
	return SessionWithDeviceType(domain.DeviceTypeTablet)
}
