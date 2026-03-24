package integration

import (
	"net/http"
	"strconv"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// ListIntegrations handles GET /integrations
func (ctrl *Controller) ListIntegrations(c *gin.Context) {
	pagination, err := httpx.GetPagination(c)
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	search := c.Query("search")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active, _ := strconv.ParseBool(activeStr)
		isActive = &active
	}

	filter := domain.IntegrationFilter{
		Limit:    int(pagination.Limit),
		Offset:   int(pagination.Offset),
		Search:   search,
		IsActive: isActive,
	}

	res, total, err := ctrl.useCase.ListIntegrations(c.Request.Context(), filter)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	meta := response.Meta{
		Total:  total,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	}

	response.ControllerResponse(c, http.StatusOK, res, meta, true)
}
