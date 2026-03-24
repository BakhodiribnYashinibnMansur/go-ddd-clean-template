package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// DeleteAPIKey handles DELETE /api-keys/:id
func (ctrl *Controller) DeleteAPIKey(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid api key id"), http.StatusBadRequest)
		return
	}

	err = ctrl.useCase.DeleteAPIKey(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusNoContent, nil, nil, true)
}
