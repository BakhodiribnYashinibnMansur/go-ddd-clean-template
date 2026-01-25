package policy

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get policy
// @Description Get policy by ID
// @Tags        authz-policies
// @Accept      json
// @Produce     json
// @Param       policy_id path string true "Policy ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/policies/{policy_id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPolicyID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - get - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid policy id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.Policy() }) {
		return
	}

	policy, err := c.u.Authz.Policy.Get(ctx.Request.Context(), &domain.PolicyFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, policy, nil, true)
}
