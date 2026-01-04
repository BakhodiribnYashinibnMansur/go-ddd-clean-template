package permission

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     List permissions
// @Description Get list of permissions with pagination and optional filtering
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       name query string false "Filter by name"
// @Success     200 {object} response.SuccessResponse{data=[]domain.Permission}
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/permissions [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := util.GetPagination(ctx)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	permName := util.GetNullStringQuery(ctx, consts.QueryName)
	filter := domain.PermissionsFilter{
		Pagination: &pagination,
	}
	if permName != "" {
		filter.Name = &permName
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGets, func(count int) any { return mock.Permissions(count) }) {
		return
	}

	perms, count, err := c.u.Authz.Permission.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, perms, count, true)
}
