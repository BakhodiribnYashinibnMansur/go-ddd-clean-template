package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetIntegration handles GET /integrations/:id
func (ctrl *Controller) GetIntegration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid integration id", nil, false)
		return
	}

	res, err := ctrl.useCase.GetIntegration(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusNotFound, "integration not found", nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
