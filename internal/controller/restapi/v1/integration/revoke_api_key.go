package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RevokeAPIKey handles POST /api-keys/:id/revoke
func (ctrl *Controller) RevokeAPIKey(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid api key id", nil, false)
		return
	}

	err = ctrl.useCase.RevokeAPIKey(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusOK, "api key revoked", nil, true)
}
