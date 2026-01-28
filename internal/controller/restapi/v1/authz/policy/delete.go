package policy

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete policy
// @Description Delete a policy by ID
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
// @Router      /authz/policies/{policy_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPolicyID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - delete - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid policy id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Policy deleted successfully") {
		return
	}

	err = c.u.Authz.Policy().Delete(ctx.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
