package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// GetIntegration handles GET /integrations/:id
func (ctrl *Controller) GetIntegration(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid integration id"), http.StatusBadRequest)
		return
	}

	res, err := ctrl.useCase.GetIntegration(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("integration not found"), http.StatusNotFound)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
