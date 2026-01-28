package session

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// UpdateActivity godoc
// @Summary     Update session activity
// @Description Update last activity timestamp
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id      path string true "Session UUID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/{id}/activity [put]
func (c *Controller) UpdateActivity(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - session - updateActivity - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid session id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "Activity updated successfully") {
		return
	}

	filter := &domain.SessionFilter{ID: &id}
	err = c.s.User.Session().UpdateActivity(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
