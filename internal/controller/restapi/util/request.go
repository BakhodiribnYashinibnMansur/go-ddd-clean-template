package util

import (
	"strconv"
	"strings"

	"gct/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request Information Helpers (Headers, IP, Domain)

func GetLanguage(ctx *gin.Context) string {
	defaultLanguage := "en"
	lang := ctx.GetHeader(consts.HeaderLanguage)
	if lang != "" {
		return strings.ToLower(lang)
	}
	return defaultLanguage
}

func GetVersion(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderAppVersion)
}

func GetUserAgent(ctx *gin.Context) string {
	return ctx.Request.UserAgent()
}

func GetDeviceID(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXDeviceID)
}

func GetDeviceIDUUID(ctx *gin.Context) uuid.UUID {
	id, err := uuid.Parse(ctx.GetHeader(consts.HeaderXDeviceID))
	if err != nil {
		return uuid.Nil
	}
	return id
}

func GetAPIKey(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXAPIKey)
}

func GetRequestID(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXRequestID)
}

func GetAuthorization(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderAuthorization)
}

func GetHeader(ctx *gin.Context, name string) string {
	return ctx.GetHeader(name)
}

func GetForwardedProto(ctx *gin.Context) string {
	return ctx.GetHeader(consts.HeaderXForwardedProto)
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
	host := ctx.Request.Header.Get(consts.HeaderOrigin)
	parts := strings.Split(host, "//")
	if len(parts) > 1 {
		return parts[1]
	}
	return host
}

func GetApiKeyType(ctx *gin.Context) (string, error) {
	apiKeyType := ctx.GetHeader(consts.HeaderXApiKeyType)
	if apiKeyType == "" {
		return "", ErrApiKeyTypeNotFound
	}
	return apiKeyType, nil
}

func ResponseHeaderXTotalCountWrite(ctx *gin.Context, total int64) {
	ctx.Writer.Header().Set(consts.HeaderXTotalCount, strconv.Itoa(int(total)))
}
