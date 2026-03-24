package ratelimit

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Update(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	var req domain.UpdateRateLimitRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	res, err := ctrl.useCase.Update(c.Request.Context(), id, req)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
