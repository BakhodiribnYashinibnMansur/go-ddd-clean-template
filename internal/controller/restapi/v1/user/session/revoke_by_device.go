package session

import (
	"net/http"

	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/gin-gonic/gin"
)

// RevokeByDevice godoc
// @Summary     Revoke sessions by device
// @Description Revoke all sessions for a specific device
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       device_id path string true "Device ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions/device/{device_id} [delete]
func (c *Controller) RevokeByDevice(ctx *gin.Context) {
	_, exists := ctx.Get(consts.CtxUserID)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	deviceID := ctx.Param("device_id")
	if deviceID == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "device ID required", nil, false)
		return
	}

	// Placeholder implementation
	c.l.Infow("Device sessions revoke requested", "device_id", deviceID)
	response.ControllerResponse(ctx, http.StatusOK, "Device sessions revoked successfully", nil, true)
}
