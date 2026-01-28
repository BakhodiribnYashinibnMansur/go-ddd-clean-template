package relation

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// AddUser godoc
// @Summary     Add user to relation
// @Description Add a user to an organizational relation
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
// @Router      /authz/relations/{relation_id}/users/{user_id} [post]
func (c *Controller) AddUser(ctx *gin.Context) {
	relationID, err := httpx.GetUUIDParam(ctx, consts.ParamRelationID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - add_user - relation uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid relation id", nil, false)
		return
	}

	userID, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - add_user - user uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "User added to relation successfully") {
		return
	}

	err = c.u.Authz.Relation().AddUser(ctx.Request.Context(), userID, relationID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
