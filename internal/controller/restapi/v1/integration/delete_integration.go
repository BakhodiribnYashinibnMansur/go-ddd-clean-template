package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteIntegration handles DELETE /integrations/:id
func (ctrl *Controller) DeleteIntegration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid integration id", nil, false)
		return
	}

	err = ctrl.useCase.DeleteIntegration(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusNoContent, nil, nil, true)
}
