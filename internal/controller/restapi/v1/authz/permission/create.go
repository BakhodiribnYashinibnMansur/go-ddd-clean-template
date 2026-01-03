package permission

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Create godoc
// @Summary     Create a new permission
// @Description Create a new permission
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       request body domain.Permission true "Permission creation body"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/permissions [post]
func (c *Controller) Create(ctx *gin.Context) {
	var perm domain.Permission
	if err := ctx.ShouldBindJSON(&perm); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	err := c.u.Authz.Permission.Create(ctx.Request.Context(), &perm)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
