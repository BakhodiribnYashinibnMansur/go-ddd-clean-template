package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/cookie"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SignOut godoc
// @Summary     Sign Out
// @Description Revoke current session
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-out [post]
func (c *Controller) SignOut(ctx *gin.Context) {
	// Require Auth Middleware to populate CtxSessionID
	sessionIDRaw, exists := ctx.Get(consts.CtxSessionID)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	sessionID, ok := sessionIDRaw.(string)
	if !ok {
		// If it's uuid.UUID, convert to string
		if uid, okUID := sessionIDRaw.(uuid.UUID); okUID {
			sessionID = uid.String()
		} else {
			util.LogError(c.l, nil, "http - v1 - auth - signout - invalid session id type")
			response.ControllerResponse(ctx, http.StatusInternalServerError, "invalid session id type", nil, false)
			return
		}
	}

	err := c.u.User.Client.SignOut(ctx.Request.Context(), uc_client.SignOutInput{
		SessionID: sessionID,
	})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	cookie.ExpireCookies(ctx, c.cfg.Cookie, consts.COOKIE_ACCESS_TOKEN, consts.COOKIE_REFRESH_TOKEN)

	response.ControllerResponse(ctx, http.StatusOK, "Signed out successfully", nil, true)
}
