package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteAPIKey handles DELETE /api-keys/:id
func (ctrl *Controller) DeleteAPIKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid api key id", nil, false)
		return
	}

	err = ctrl.useCase.DeleteAPIKey(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusNoContent, nil, nil, true)
}
