package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// RevokeAll godoc
// @Summary     Revoke all other sessions
// @Description Revoke all user sessions except the current one (Currently revokes current one as placeholder)
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions/revoke-all [post]
func (c *Controller) RevokeAll(ctx *gin.Context) {
	userID, err := util.GetUserIDUUID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	sid, err := util.GetCtxSessionID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session not found", nil, false)
		return
	}

	sessionUUID, err := uuid.Parse(sid)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - session - revoke_all - uuid")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "invalid session ID format", nil, false)
		return
	}

	// Just revoke current session for now
	filter := &domain.SessionFilter{ID: &sessionUUID}
	err = c.s.User.Session.Revoke(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	c.l.Infow("Session revoked for user", "user_id", userID)
	response.ControllerResponse(ctx, http.StatusOK, "Session revoked successfully", nil, true)
}
