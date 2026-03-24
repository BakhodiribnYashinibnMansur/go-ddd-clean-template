package session

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new session
// @Description Create a session for user
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       request body domain.Session true "Session creation query"
// @Success     201 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions [post]
func (c *Controller) Create(ctx *gin.Context) {
	var req domain.Session
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpx.LogError(c.l, err, "http - v1 - session - create - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.Session() }) {
		return
	}

// Using pointer for session as requested
	createSession, err := c.s.User.Session().Create(ctx.Request.Context(), &req)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, createSession, nil, true)
}
