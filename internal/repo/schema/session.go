package schema

// Table name
const TableSession = "session"

// Session table columns
const (
	SessionID               = "id"
	SessionUserID           = "user_id"
	SessionDeviceID         = "device_id"
	SessionDeviceName       = "device_name"
	SessionDeviceType       = "device_type"
	SessionIPAddress        = "ip_address"
	SessionUserAgent        = "user_agent"
	SessionOS               = "os"
	SessionOSVersion        = "os_version"
	SessionBrowser          = "browser"
	SessionBrowserVersion   = "browser_version"
	SessionFCMToken         = "fcm_token"
	SessionData             = "data"
	SessionRefreshTokenHash = "refresh_token_hash"
	SessionExpiresAt        = "expires_at"
	SessionLastActivity     = "last_activity"
	SessionRevoked          = "revoked"
	SessionCreatedAt        = "created_at"
	SessionUpdatedAt        = "updated_at"
)
