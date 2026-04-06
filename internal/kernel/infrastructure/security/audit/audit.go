package audit

import (
	"context"

	"github.com/google/uuid"
)

// Event constants.
const (
	EventSignInSuccess    = "sign_in_success"
	EventSignInFailed     = "sign_in_failed"
	EventRefreshSuccess   = "refresh_success"
	EventRefreshReuse     = "refresh_reuse_detected"
	EventAPIKeyMismatch   = "api_key_mismatch"
	EventCrossIntegration = "cross_integration_attempt"
	EventSessionRevoked   = "session_revoked"
	EventAccountLocked    = "account_locked"
	EventTBHMismatch      = "tbh_mismatch"
	EventKeyGenerated     = "key_generated"
	EventKeyRotated       = "key_rotated"
	EventAPIKeyScraping   = "api_key_scraping_detected"
)

// Entry is one audit record.
type Entry struct {
	Event           string
	IntegrationName string
	UserID          *uuid.UUID     // nil for anonymous events (bad API key)
	SessionID       *uuid.UUID
	IPAddress       string
	UserAgent       string
	Metadata        map[string]any // free-form context
}

// Logger writes audit entries. Implementations must be safe for concurrent use.
type Logger interface {
	Log(ctx context.Context, entry Entry)
}
