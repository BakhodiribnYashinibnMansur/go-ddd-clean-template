package cookie

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/config"
	"gct/consts"
)

// SaveCookies saves multiple cookies using settings from config.
func SaveCookies(ctx *gin.Context, data map[string]string, cfg config.Cookie) {
	for k, v := range data {
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     k,
			Value:    v,
			MaxAge:   cfg.MaxAge,
			Path:     consts.CookiePath,
			Domain:   cfg.Domain,
			Secure:   cfg.IsSecure(),
			HttpOnly: cfg.IsHttpOnly(),
			SameSite: http.SameSiteLaxMode,
		})
	}
}

// GetCookie retrieves a cookie value from the request.
func GetCookie(ctx *gin.Context, key string) string {
	data, err := ctx.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return data.Value
}

// ExpireCookies expires specified cookies using settings from config and consts.
func ExpireCookies(ctx *gin.Context, cfg config.Cookie, keys ...string) {
	for _, key := range keys {
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     key,
			Value:    "",
			MaxAge:   -1,
			Path:     consts.CookiePath,
			Domain:   cfg.Domain,
			Secure:   cfg.IsSecure(),
			HttpOnly: cfg.IsHttpOnly(),
			SameSite: http.SameSiteLaxMode,
		})
	}
}

// GetCookieConfig returns a cookie configuration based on constants.
func GetCookieConfig(key, value string) *http.Cookie {
	return &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   consts.CookieExpiredTime,
		Path:     consts.CookiePath,
		Domain:   consts.CookieDomain,
		Secure:   true,
		HttpOnly: consts.CookieHttpOnly,
		SameSite: http.SameSiteNoneMode,
	}
}
