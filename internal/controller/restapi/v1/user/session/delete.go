package session

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete/Revoke session
// @Description Revoke a session by ID
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id  path string true "Session UUID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/{id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - session - delete - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid session id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Session revoked successfully") {
		return
	}

	filter := &domain.SessionFilter{ID: &id}
	err = c.s.User.Session.Delete(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
