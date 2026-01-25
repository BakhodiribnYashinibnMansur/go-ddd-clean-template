package policy

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary     Update policy
// @Description Update an existing policy
// @Tags        authz-policies
// @Accept      json
// @Produce     json
// @Param       policy_id path string true "Policy ID"
// @Param       request body domain.Policy true "Policy update body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/policies/{policy_id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPolicyID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - update - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid policy id", nil, false)
		return
	}

	var policy domain.Policy
	if err := ctx.ShouldBindJSON(&policy); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - update - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	policy.ID = id

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "Policy updated successfully") {
		return
	}

	err = c.u.Authz.Policy.Update(ctx.Request.Context(), &policy)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
