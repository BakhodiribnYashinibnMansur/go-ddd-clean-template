package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// CreateAPIKey handles POST /integrations/:id/keys
func (ctrl *Controller) CreateAPIKey(c *gin.Context) {
	integrationID, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid integration id"), http.StatusBadRequest)
		return
	}

	var req domain.CreateAPIKeyRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	req.IntegrationID = integrationID

	res, rawKey, err := ctrl.useCase.CreateAPIKey(c.Request.Context(), req)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	// For security, raw key is only returned once during creation
	data := map[string]any{
		"api_key": res,
		"raw_key": rawKey,
	}

	response.ControllerResponse(c, http.StatusCreated, data, nil, true)
}
