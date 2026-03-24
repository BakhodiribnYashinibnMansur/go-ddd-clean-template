package dataexport

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) List(c *gin.Context) {
	pagination, err := httpx.GetPagination(c)
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	filter := domain.DataExportFilter{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Limit:  int(pagination.Limit),
		Offset: int(pagination.Offset),
	}
	items, total, err := ctrl.useCase.List(c.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, items, response.Meta{Total: total, Limit: pagination.Limit, Offset: pagination.Offset}, true)
}
