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

// SignOut godoc
// @Summary     Sign Out
// @Description Revoke current session
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.SignOutIn true "Sign out request body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-out [post]
func (c *Controller) SignOut(ctx *gin.Context) {
	userId, err := util.GetUserID(ctx)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusUnauthorized)
		return
	}

	var req domain.SignOutIn
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.LogError(c.l, err, "http - v1 - auth - signout - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	req.UserID = userId

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeUpdate, "Signed out successfully") {
		return
	}

	err = c.u.User.Client.SignOut(ctx.Request.Context(), &req)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	cookie.ExpireCookies(ctx, c.cfg.Cookie, consts.COOKIE_ACCESS_TOKEN, consts.COOKIE_REFRESH_TOKEN)

	response.ControllerResponse(ctx, http.StatusOK, "Signed out successfully", nil, true)
}
