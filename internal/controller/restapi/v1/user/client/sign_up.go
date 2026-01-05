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

// SignUp godoc
// @Summary     Sign Up
// @Description Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "User info"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-up [post]
func (c *Controller) SignUp(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		util.LogError(c.l, err, "http - v1 - auth - signup - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return mock.SignInOut() }) {
		return
	}

	out, err := c.u.User.Client.SignUp(ctx.Request.Context(), &domain.SignUpIn{
		Username:  username,
		Phone:     *user.Phone,
		Password:  user.Password,
		DeviceID:  util.GetDeviceIDUUID(ctx),
		UserAgent: util.GetUserAgent(ctx),
		IP:        util.GetIPAddress(ctx),
	})
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	// Set Cookies
	accessCookieCfg := c.cfg.Cookie
	accessCookieCfg.MaxAge = int(c.cfg.JWT.AccessTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_ACCESS_TOKEN: out.AccessToken}, accessCookieCfg)

	refreshCookieCfg := c.cfg.Cookie
	refreshCookieCfg.MaxAge = int(c.cfg.JWT.RefreshTTL.Seconds())
	cookie.SaveCookies(ctx, map[string]string{consts.COOKIE_REFRESH_TOKEN: out.RefreshToken}, refreshCookieCfg)

	response.ControllerResponse(ctx, http.StatusCreated, out, nil, true)
}
