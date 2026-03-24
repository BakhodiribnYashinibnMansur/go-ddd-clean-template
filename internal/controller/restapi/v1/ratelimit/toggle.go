package ratelimit

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Toggle(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	res, err := ctrl.useCase.Toggle(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
