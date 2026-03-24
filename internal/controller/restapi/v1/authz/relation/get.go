package relation

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get relation
// @Description Get relation by ID
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       relation_id path string true "Relation ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations/{relation_id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamRelationID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - get - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid relation id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.Relation() }) {
		return
	}

	relation, err := c.u.Authz.Relation().Get(ctx.Request.Context(), &domain.RelationFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, relation, nil, true)
}
