package integration

import (
	"net/http"
	"strconv"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

// ListIntegrations handles GET /integrations
func (ctrl *Controller) ListIntegrations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active, _ := strconv.ParseBool(activeStr)
		isActive = &active
	}

	filter := domain.IntegrationFilter{
		Limit:    limit,
		Offset:   offset,
		Search:   search,
		IsActive: isActive,
	}

	res, total, err := ctrl.useCase.ListIntegrations(c.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	meta := response.Meta{
		Total:  total,
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	response.ControllerResponse(c, http.StatusOK, res, meta, true)
}
