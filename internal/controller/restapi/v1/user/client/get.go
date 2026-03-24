package client

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get user by ID
// @Description Retrieve a user's details by their unique identifier
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id  path string true "User ID (UUID format)" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid user ID format"
// @Failure     401 {object} response.ErrorResponse "Unauthorized - missing or invalid token"
// @Failure     403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure     404 {object} response.ErrorResponse "User not found"
// @Failure     500 {object} response.ErrorResponse "Internal server error"
// @Security    BearerAuth
// @Router      /users/{user_id} [get]
func (c *Controller) User(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - get - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.User() }) {
		return
	}

	out, err := c.u.User.Client().Get(ctx.Request.Context(), &domain.UserFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
