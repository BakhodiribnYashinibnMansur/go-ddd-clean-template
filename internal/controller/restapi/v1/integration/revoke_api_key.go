package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// RevokeAPIKey handles POST /api-keys/:id/revoke
func (ctrl *Controller) RevokeAPIKey(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid api key id"), http.StatusBadRequest)
		return
	}

	err = ctrl.useCase.RevokeAPIKey(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusOK, "api key revoked", nil, true)
}
