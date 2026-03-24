package policy

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gets godoc
// @Summary     List policies
// @Description Get list of policies with filtering
// @Tags        authz-policies
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       permission_id query string false "Permission ID"
// @Param       active query bool false "Active status"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/policies [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.PoliciesFilter{
		Pagination: &pagination,
		PolicyFilter: domain.PolicyFilter{
			PermissionID: func() *uuid.UUID {
				idStr := httpx.GetNullStringQuery(ctx, "permission_id")
				if idStr != "" {
					uid, err := uuid.Parse(idStr)
					if err == nil {
						return &uid
					}
				}
				return nil
			}(),
			Active: func() *bool {
				activeStr := httpx.GetNullStringQuery(ctx, "active")
				switch activeStr {
				case "true":
					t := true
					return &t
				case "false":
					f := false
					return &f
				}
				return nil
			}(),
		},
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Policies(count) }) {
		return
	}

	policies, count, err := c.u.Authz.Policy().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, policies, count, true)
}
