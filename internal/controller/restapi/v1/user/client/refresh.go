package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// RefreshToken godoc
// @Summary     Refresh tokens
// @Description Rotates access and refresh tokens using a valid refresh token.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.RefreshIn true "Refresh request body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/refresh [post]
func (c *Controller) RefreshToken(ctx *gin.Context) {
	var req domain.RefreshIn
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.LogError(c.l, err, "http - v1 - auth - refresh - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Rotate session
	out, err := c.u.User.Client.RotateSession(ctx.Request.Context(), &req)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid refresh token or session", nil, false)
		return
	}

	// Set new Cookies
	accessCookieCfg := c.cfg.Cookie
	accessCookieCfg.MaxAge = int(c.cfg.JWT.AccessTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_ACCESS_TOKEN: out.AccessToken}, accessCookieCfg)

	refreshCookieCfg := c.cfg.Cookie
	refreshCookieCfg.MaxAge = int(c.cfg.JWT.RefreshTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_REFRESH_TOKEN: out.RefreshToken}, refreshCookieCfg)

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
