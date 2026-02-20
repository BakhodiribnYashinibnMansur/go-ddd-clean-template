package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListAPIKeys handles GET /integrations/:id/keys
func (ctrl *Controller) ListAPIKeys(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid integration id", nil, false)
		return
	}

	res, err := ctrl.useCase.ListAPIKeys(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
