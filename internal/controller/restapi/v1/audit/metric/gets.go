package metric

import (
	"net/http"
	"time"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     Get function metrics
// @Description Retrieve function execution metrics
// @Tags        metrics
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       name query string false "Function Name"
// @Param       is_panic query bool false "Is Panic"
// @Param       from_date query string false "From Date (RFC3339)"
// @Param       to_date query string false "To Date (RFC3339)"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /metrics/functions [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - metric - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.FunctionMetricsFilter{
		Pagination: &pagination,
		Name: func() *string {
			s := httpx.GetNullStringQuery(ctx, "name")
			if s == "" {
				return nil
			}
			return &s
		}(),
	}

	isPanicStr := httpx.GetNullStringQuery(ctx, "is_panic")
	if isPanicStr == "true" {
		t := true
		filter.IsPanic = &t
	} else if isPanicStr == "false" {
		f := false
		filter.IsPanic = &f
	}

	fromDateStr := httpx.GetNullStringQuery(ctx, "from_date")
	if fromDateStr != "" {
		t, err := time.Parse(time.RFC3339, fromDateStr)
		if err == nil {
			filter.FromDate = &t
		}
	}

	toDateStr := httpx.GetNullStringQuery(ctx, "to_date")
	if toDateStr != "" {
		t, err := time.Parse(time.RFC3339, toDateStr)
		if err == nil {
			filter.ToDate = &t
		}
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.FunctionMetrics(count) }) {
		return
	}

	metrics, total, err := c.u.Audit.Metric().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, metrics, total, true)
}
