package session

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// GetActiveSessions godoc
// @Summary     Get active sessions
// @Description List all active sessions for the current user
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       offset  query int false "Offset"
// @Param       limit   query int false "Limit"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions [get]
func (c *Controller) Sessions(ctx *gin.Context) {
	userID, err := httpx.GetUserID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	filter := &domain.SessionsFilter{
		SessionFilter: domain.SessionFilter{
			UserID: &userID,
		},
		Pagination: &pagination,
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Sessions(count) }) {
		return
	}

	sessions, total, err := c.s.User.Session().Gets(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	meta := &response.Meta{
		Total:  int64(total),
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Page:   pagination.Offset/pagination.Limit + 1,
	}

	response.ControllerResponse(ctx, http.StatusOK, sessions, meta, true)
}
