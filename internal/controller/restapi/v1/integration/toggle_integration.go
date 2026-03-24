package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// ToggleIntegration handles POST /integrations/:id/toggle
func (ctrl *Controller) ToggleIntegration(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid integration id"), http.StatusBadRequest)
		return
	}

	res, err := ctrl.useCase.ToggleIntegration(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
