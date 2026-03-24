// Package util provides cross-cutting helper functions for the Gin REST API controllers.
package httpx

import (
	"strconv"
	"strings"

	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request Information Helpers extract metadata from HTTP headers and properties.

// GetLanguage identifies the preferred content language from the request headers.
// Defaults to "en" if the header is missing.
func GetLanguage(ctx *gin.Context) string {
	defaultLanguage := "en"
	lang := ctx.GetHeader(consts.HeaderLanguage)
	if lang != "" {
		return strings.ToLower(lang)
	}
	return defaultLanguage
}

// GetVersion retrieves the application version reported by the client.
func GetVersion(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderAppVersion)
}

// GetUserAgent extracts the raw User-Agent string from the request.
func GetUserAgent(ctx *gin.Context) string {
	return ctx.Request.UserAgent()
}

// GetDeviceID retrieves the unique device identifier from custom headers.
func GetDeviceID(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXDeviceID)
}

// GetDeviceIDUUID attempts to parse the device ID header as a UUID.
// Returns uuid.Nil if parsing fails.
func GetDeviceIDUUID(ctx *gin.Context) uuid.UUID {
	id, err := uuid.Parse(ctx.GetHeader(consts.HeaderXDeviceID))
	if err != nil {
		return uuid.Nil
	}
	return id
}

// GetAPIKey extracts the specialized API key from security headers.
func GetAPIKey(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXAPIKey)
}

// GetRequestID retrieves the tracking identifier for the current request.
func GetCtxRequestID(ctx *gin.Context) string {
	requestId := ctx.GetHeader(consts.HeaderXRequestID)
	if requestId == "" {
		return uuid.New().String()
	}
	return requestId
}

// GetAuthorization extracts the raw Authorization header content (e.g. Bearer token).
func GetAuthorization(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderAuthorization)
}

// GetHeader is a generic wrapper to retrieve any header value by name.
func GetHeader(ctx *gin.Context, name string) string {
	return ctx.GetHeader(name)
}

// GetForwardedProto determines if the request was forwarded as HTTP or HTTPS.
func GetForwardedProto(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXForwardedProto)
}

// GetIPAddress resolves the client's IP address, handling common localhost aliases.
func GetIPAddress(ctx *gin.Context) string {
	localhostV6 := "::1"
	localhostV4 := "127.0.0.1"
	ip := ctx.ClientIP()
	if ip == localhostV6 {
		return localhostV4
	}
	return ip
}

// GetClientDomain parses the host or origin header to identify the client's domain.
func GetClientDomain(ctx *gin.Context) string {
	host := ctx.Request.Header.Get(consts.HeaderOrigin)
	parts := strings.Split(host, "//")
	if len(parts) > 1 {
		return parts[1]
	}
	return host
}

// GetApiKeyType identifies the category or level of the provided API key.
func GetApiKeyType(ctx *gin.Context) (string, error) {
	apiKeyType := ctx.GetHeader(consts.HeaderXApiKeyType)
	if apiKeyType == "" {
		return "", ErrApiKeyTypeNotFound
	}
	return apiKeyType, nil
}

// ResponseHeaderXTotalCountWrite sets the X-Total-Count header, useful for client-side pagination.
func ResponseHeaderXTotalCountWrite(ctx *gin.Context, total int64) {
	ctx.Writer.Header().Set(consts.HeaderXTotalCount, strconv.Itoa(int(total)))
}
