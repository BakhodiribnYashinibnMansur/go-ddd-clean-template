package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Users godoc
// @Summary     Get users
// @Description Retrieve users with pagination
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       phone query string false "Phone"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /users [get]
func (c *Controller) Users(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - users - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.UsersFilter{
		UserFilter: domain.UserFilter{
			Phone: func() *string {
				p := httpx.GetNullStringQuery(ctx, consts.QueryPhone)
				if p == "" {
					return nil
				}
				return &p
			}(),
		},
		Pagination: &pagination,
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Users(count) }) {
		return
	}

	users, total, err := c.u.User.Client.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, users, total, true)
}
