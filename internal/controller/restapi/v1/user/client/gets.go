package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/domain/mock"
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
// @Success     200 {object} response.SuccessResponse{data=[]domain.User}
// @Failure     400 {object} response.ErrorResponse
// @Router      /users [get]
func (c *Controller) Users(ctx *gin.Context) {
	pagination, err := util.GetPagination(ctx)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - client - users - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.UsersFilter{
		UserFilter: domain.UserFilter{
			Phone: func() *string {
				p := util.GetNullStringQuery(ctx, consts.QueryPhone)
				if p == "" {
					return nil
				}
				return &p
			}(),
		},
		Pagination: &pagination,
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGets, func(count int) any { return mock.Users(count) }) {
		return
	}

	users, total, err := c.u.User.Client.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, users, total, true)
}
