package sitesetting

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination params", nil, false)
		return
	}

	filter := domain.SiteSettingsFilter{
		Pagination: &pagination,
	}

	if category := ctx.Query("category"); category != "" {
		filter.Category = &category
	}

	settings, total, err := c.uc.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	meta := response.Meta{
		Total:  int64(total),
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	}

	response.ControllerResponse(ctx, http.StatusOK, settings, meta, true)
}
