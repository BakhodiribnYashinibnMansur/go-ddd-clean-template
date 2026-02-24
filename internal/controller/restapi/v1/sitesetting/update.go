package sitesetting

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

type updateRequest struct {
	Value string `json:"value" binding:"required"`
}

func (c *Controller) UpdateByKey(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "key is required", nil, false)
		return
	}

	var req updateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	if err := c.uc.UpdateByKey(ctx.Request.Context(), key, req.Value); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "Setting updated", nil, true)
}
