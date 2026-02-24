package policy

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Toggle godoc
// @Summary     Toggle policy active state
// @Description Flip the active boolean on a policy
// @Tags        authz-policies
// @Accept      json
// @Produce     json
// @Param       policy_id path string true "Policy ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/policies/{policy_id}/toggle [post]
func (c *Controller) Toggle(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPolicyID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - toggle - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid policy id", nil, false)
		return
	}

	if err := c.u.Authz.Policy().Toggle(ctx.Request.Context(), id); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
