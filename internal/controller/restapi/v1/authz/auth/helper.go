package auth

import (
	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// populateSessionInfo extracts metadata from the request context to fill session details.
func (c *Controller) populateSessionInfo(ctx *gin.Context, s *domain.SessionIn) {
	if s == nil {
		return
	}
	s.DeviceID = httpx.GetDeviceIDUUID(ctx)
	s.UserAgent = httpx.GetUserAgent(ctx)
	s.IP = httpx.GetIPAddress(ctx)
}

// setAuthCookies persists access and refresh tokens into secure, HTTP-only cookies.
func (c *Controller) setAuthCookies(ctx *gin.Context, out *domain.SignInOut) {
	accessCookieCfg := c.cfg.Cookie
	accessCookieCfg.MaxAge = int(c.cfg.JWT.AccessTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_ACCESS_TOKEN: out.AccessToken}, accessCookieCfg)

	refreshCookieCfg := c.cfg.Cookie
	refreshCookieCfg.MaxAge = int(c.cfg.JWT.RefreshTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_REFRESH_TOKEN: out.RefreshToken}, refreshCookieCfg)
}
