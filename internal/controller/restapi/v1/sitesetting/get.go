package sitesetting

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetByKey(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "key is required", nil, false)
		return
	}

	setting, err := c.uc.GetByKey(ctx.Request.Context(), key)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, setting, nil, true)
}
