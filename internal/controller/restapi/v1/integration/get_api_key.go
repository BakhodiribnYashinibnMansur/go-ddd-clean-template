package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// GetAPIKey handles GET /api-keys/:id
func (ctrl *Controller) GetAPIKey(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid api key id"), http.StatusBadRequest)
		return
	}

	res, err := ctrl.useCase.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("api key not found"), http.StatusNotFound)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
