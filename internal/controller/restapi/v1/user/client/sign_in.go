package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// SignIn godoc
// @Summary     Sign In
// @Description Authenticate user and return tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "Credentials"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-in [post]
func (c *Controller) SignIn(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		util.LogError(c.l, err, "http - v1 - auth - signin - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return mock.SignInOut() }) {
		return
	}

	out, err := c.u.User.Client.SignIn(ctx.Request.Context(), &domain.SignInIn{
		Phone:     phone,
		Password:  user.Password,
		DeviceID:  util.GetDeviceIDUUID(ctx),
		UserAgent: util.GetUserAgent(ctx),
		IP:        util.GetIPAddress(ctx),
	})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid credentials", nil, false)
		return
	}

	// Set Cookies
	accessCookieCfg := c.cfg.Cookie
	accessCookieCfg.MaxAge = int(c.cfg.JWT.AccessTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_ACCESS_TOKEN: out.AccessToken}, accessCookieCfg)

	refreshCookieCfg := c.cfg.Cookie
	refreshCookieCfg.MaxAge = int(c.cfg.JWT.RefreshTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_REFRESH_TOKEN: out.RefreshToken}, refreshCookieCfg)

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
