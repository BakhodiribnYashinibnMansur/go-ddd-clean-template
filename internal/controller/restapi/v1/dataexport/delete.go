package dataexport

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.ControllerResponse(c, http.StatusBadRequest, "id is required", nil, false)
		return
	}
	if err := ctrl.useCase.Delete(c.Request.Context(), id); err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, nil, nil, true)
}
