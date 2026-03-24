package consts

const (
	// BASE URL
	BaseURLV1 string = "/api/v1"
	// COOKIE KEY
	CookiePath         string = "/"
	CookieAccessToken  string = "c_at"
	CookieRefreshToken string = "c_rt"
	CookieUserID       string = "c_uid"
	CookiePlatformType string = "c_pt"
	CookieUserFullname string = "c_ufn"
	CookieUserPhone    string = "c_uph"
	CookieUserRoleID   string = "c_uro"
	CookieCsrfToken    string = "c_csrf"

	CookieExpiredTime int    = 3600
	CookieDomain      string = "localhost"
	CookieHttpOnly    bool   = true

	// TELEGRAM
	TelegramErrorTopicID string = "2"

	// TOKEN TYPE
	TokenTypeAccess  string = "access"
	TokenTypeRefresh string = "refresh"

	// API KEY
	DefaultAPIKeyPrefix string = "sk_live"
	APIKeyCharset       string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	APIKeyLength        int    = 32
)
