package relation

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary     Update relation
// @Description Update an existing relation
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       relation_id path string true "Relation ID"
// @Param       request body domain.Relation true "Relation update body"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations/{relation_id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamRelationID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - update - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid relation id", nil, false)
		return
	}

	var relation domain.Relation
	if err := ctx.ShouldBindJSON(&relation); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - update - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	relation.ID = id

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "Relation updated successfully") {
		return
	}

	err = c.u.Authz.Relation().Update(ctx.Request.Context(), &relation)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
