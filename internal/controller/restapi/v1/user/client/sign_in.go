package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/cookie"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
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

	out, err := c.u.User.Client.SignIn(ctx.Request.Context(), uc_client.SignInInput{
		Phone:     user.Phone,
		Password:  user.Password,
		DeviceID:  ctx.GetHeader("X-Device-ID"),
		UserAgent: ctx.GetHeader("User-Agent"),
		IP:        ctx.ClientIP(),
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
