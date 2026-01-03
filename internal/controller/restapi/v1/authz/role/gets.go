package role

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Gets godoc
// @Summary     List roles
// @Description Get list of roles with pagination and optional filtering
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       name query string false "Filter by name"
// @Success     200 {object} response.SuccessResponse{data=[]domain.Role}
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := util.GetPagination(ctx)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	roleName := util.GetNullStringQuery(ctx, consts.QueryName)
	filter := domain.RolesFilter{
		Pagination: &pagination,
	}
	if roleName != "" {
		filter.Name = &roleName
	}

	roles, count, err := c.u.Authz.Role.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, roles, count, true)
}
