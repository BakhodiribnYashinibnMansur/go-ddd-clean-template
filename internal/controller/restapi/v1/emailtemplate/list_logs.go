package emailtemplate

import (
	"net/http"
	"strconv"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) ListLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter := domain.EmailLogFilter{
		TemplateID: c.Query("template_id"),
		Status:     c.Query("status"),
		Limit:      limit,
		Offset:     offset,
	}
	items, total, err := ctrl.useCase.ListLogs(c.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, items, response.Meta{Total: total, Limit: int64(limit), Offset: int64(offset)}, true)
}
