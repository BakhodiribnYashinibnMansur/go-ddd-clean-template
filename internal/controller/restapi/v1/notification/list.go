package notification

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
	filter := domain.NotificationFilter{
		Search: c.Query("search"),
		Type:   c.Query("type"),
		Limit:  int(pagination.Limit),
		Offset: int(pagination.Offset),
	}
	items, total, err := ctrl.useCase.List(c.Request.Context(), filter)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}
	response.ControllerResponse(c, http.StatusOK, items, response.Meta{Total: total, Limit: pagination.Limit, Offset: pagination.Offset}, true)
}
