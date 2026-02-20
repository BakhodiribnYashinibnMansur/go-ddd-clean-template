package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAPIKey handles GET /api-keys/:id
func (ctrl *Controller) GetAPIKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid api key id", nil, false)
		return
	}

	res, err := ctrl.useCase.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusNotFound, "api key not found", nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
