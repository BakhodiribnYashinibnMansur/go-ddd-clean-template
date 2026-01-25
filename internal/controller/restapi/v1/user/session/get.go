package session

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get session by ID
// @Description Retrieve a session details
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id  path string true "Session UUID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/{id} [get]
func (c *Controller) Session(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - session - get - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid session id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.Session() }) {
		return
	}

	filter := &domain.SessionFilter{ID: &id}
	session, err := c.s.User.Session.Get(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, session, nil, true)
}
