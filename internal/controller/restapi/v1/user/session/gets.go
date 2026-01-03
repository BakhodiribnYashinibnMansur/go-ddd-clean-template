package session

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
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
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions [get]
func (c *Controller) Sessions(ctx *gin.Context) {
	userID, err := util.GetUserIDUUID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	pagination, err := util.GetPagination(ctx)
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

	sessions, total, err := c.s.User.Session.Gets(ctx.Request.Context(), filter)
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
