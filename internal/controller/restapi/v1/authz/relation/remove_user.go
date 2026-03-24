package relation

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// RemoveUser godoc
// @Summary     Remove user from relation
// @Description Remove a user from an organizational relation
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       relation_id path string true "Relation ID"
// @Param       user_id path string true "User ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations/{relation_id}/users/{user_id} [delete]
func (c *Controller) RemoveUser(ctx *gin.Context) {
	relationID, err := httpx.GetUUIDParam(ctx, consts.ParamRelationID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - remove_user - relation uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid relation id", nil, false)
		return
	}

	userID, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - remove_user - user uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "User removed from relation successfully") {
		return
	}

	err = c.u.Authz.Relation().RemoveUser(ctx.Request.Context(), userID, relationID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
