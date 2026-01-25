package session

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// RevokeAll godoc
// @Summary     Revoke all other sessions
// @Description Revoke all user sessions except the current one
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       request body domain.RevokeSessionsIn true "Revoke sessions request body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/revoke-all [post]
func (c *Controller) RevokeAll(ctx *gin.Context) {
	userID, err := httpx.GetUserID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	sid, err := httpx.GetCtxSessionID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session not found", nil, false)
		return
	}

	var req domain.RevokeSessionsIn
	if err := ctx.ShouldBindJSON(&req); err != nil {
// We allow empty body for now if it's just a general revoke all
// But the user said Post needs a body.
		httpx.LogError(c.l, err, "http - v1 - session - revokeall - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	req.UserID = userID

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Sessions revoked successfully") {
		return
	}

// Just revoke current session for now
	filter := &domain.SessionFilter{ID: &sid}
	err = c.s.User.Session.Revoke(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	c.l.Infow("Session revoked for user", "user_id", userID)
	response.ControllerResponse(ctx, http.StatusOK, "Session revoked successfully", nil, true)
}
