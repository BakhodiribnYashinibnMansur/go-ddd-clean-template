package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Request Information Helpers (Headers, IP, Domain)

func GetLanguage(ctx *gin.Context) string {
	defaultLanguage := "en"
	lang := ctx.GetHeader(Language)
	if lang != "" {
		return strings.ToLower(lang)
	}
	return defaultLanguage
}

func GetVersion(ctx *gin.Context) string {
	return ctx.GetHeader(AppVersionHeader)
}

func GetUserAgent(ctx *gin.Context) string {
	return ctx.Request.UserAgent()
}

func GetIPAddress(ctx *gin.Context) string {
	localhostV6 := "::1"
	localhostV4 := "127.0.0.1"
	ip := ctx.ClientIP()
	if ip == localhostV6 {
		return localhostV4
	}
	return ip
}

func GetClientDomain(ctx *gin.Context) string {
	host := ctx.Request.Header.Get("Origin")
	parts := strings.Split(host, "//")
	if len(parts) > 1 {
		return parts[1]
	}
	return host
}

func GetApiKeyType(ctx *gin.Context) (string, error) {
	apiKeyType := ctx.GetHeader(ApiKeyTypeHeader)
	if apiKeyType == "" {
		return "", fmt.Errorf("%s not found", ApiKeyTypeHeader)
	}
	return apiKeyType, nil
}

func ResponseHeaderXTotalCountWrite(ctx *gin.Context, total int64) {
	ctx.Writer.Header().Set("X-Total-Count", strconv.Itoa(int(total)))
}
