package permission

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
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
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/permissions [post]
func (c *Controller) Create(ctx *gin.Context) {
	var perm domain.Permission
	if err := ctx.ShouldBindJSON(&perm); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - permission - create - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "Permission created successfully") {
		return
	}

	err := c.u.Authz.Permission().Create(ctx.Request.Context(), &perm)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
