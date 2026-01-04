package role

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new role
// @Description Create a new role with name
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       request body domain.Role true "Role creation body"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles [post]
func (c *Controller) Create(ctx *gin.Context) {
	var role domain.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeCreate, "Role created successfully") {
		return
	}

	err := c.u.Authz.Role.Create(ctx.Request.Context(), &role)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
