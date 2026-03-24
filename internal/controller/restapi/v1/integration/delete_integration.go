package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// DeleteIntegration handles DELETE /integrations/:id
func (ctrl *Controller) DeleteIntegration(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid integration id"), http.StatusBadRequest)
		return
	}

	err = ctrl.useCase.DeleteIntegration(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusNoContent, nil, nil, true)
}
