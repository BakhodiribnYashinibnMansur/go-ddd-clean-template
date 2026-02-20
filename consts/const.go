package consts

const (
	// BASE URL
	BASE_URL_V1 string = "/api/v1"
	// COOKIE KEY
	CookiePath           string = "/"
	COOKIE_ACCESS_TOKEN  string = "c_at"
	COOKIE_REFRESH_TOKEN string = "c_rt"
	COOKIE_USER_ID       string = "c_uid"
	COOKIE_PLATFORM_TYPE string = "c_pt"
	COOKIE_USER_FULLNAME string = "c_ufn"
	COOKIE_USER_PHONE    string = "c_uph"
	COOKIE_USER_ROLE_ID  string = "c_uro"
	COOKIE_CSRF_TOKEN    string = "c_csrf"

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
