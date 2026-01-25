package relation

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete relation
// @Description Delete a relation by ID
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       relation_id path string true "Relation ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations/{relation_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamRelationID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - delete - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid relation id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Relation deleted successfully") {
		return
	}

	err = c.u.Authz.Relation.Delete(ctx.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
