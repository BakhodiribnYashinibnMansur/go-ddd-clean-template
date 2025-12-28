package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
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
				p := util.GetNullStringQuery(ctx, "phone")
				if p == "" {
					return nil
				}
				return &p
			}(),
		},
		Pagination: &pagination,
	}

	out, err := c.u.User.Client.Users(ctx.Request.Context(), uc_client.UsersInput{Filter: filter})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, out.Users, pagination, true)
}
