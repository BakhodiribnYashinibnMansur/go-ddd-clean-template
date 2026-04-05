// Package cookie provides utilities for managing HTTP cookies within the Gin web framework,
// supporting consistent configuration for security attributes like HttpOnly, Secure, and SameSite.
package cookie

import (
	"net/http"

	"gct/config"
	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
)

// SaveCookies persists a map of cookie keys and values to the HTTP response.
// It applies security and lifecycle settings defined in the application's cookie configuration.
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

// GetCookie safely extracts a cookie's value from the incoming request.
// Returns an empty string if the cookie does not exist or an error occurs.
func GetCookie(ctx *gin.Context, key string) string {
	data, err := ctx.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return data.Value
}

// ExpireCookies removes specified cookies from the client by setting their MaxAge to -1.
// This effectively instructs the browser to delete the cookies immediately.
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

// GetCookieConfig instantiates a new http.Cookie with default application settings.
// Useful for creating specialized cookies that don't follow the global config.
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
