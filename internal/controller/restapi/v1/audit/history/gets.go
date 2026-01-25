package history

import (
	"net/http"
	"time"

	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gets godoc
// @Summary     Get endpoint histories
// @Description Retrieve endpoint histories with filtering and pagination
// @Tags        audit
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       user_id query string false "User ID"
// @Param       method query string false "Method"
// @Param       path query string false "Path"
// @Param       status_code query int false "Status Code"
// @Param       from_date query string false "From Date (RFC3339)"
// @Param       to_date query string false "To Date (RFC3339)"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /audit/history [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - audit - history - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.EndpointHistoriesFilter{
		Pagination: &pagination,
		EndpointHistoryFilter: domain.EndpointHistoryFilter{
			Method: func() *string {
				s := httpx.GetNullStringQuery(ctx, "method")
				if s == "" {
					return nil
				}
				return &s
			}(),
			Path: func() *string {
				s := httpx.GetNullStringQuery(ctx, "path")
				if s == "" {
					return nil
				}
				return &s
			}(),
		},
	}

	userIDStr := httpx.GetNullStringQuery(ctx, "user_id")
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &uid
		}
	}

	statusCode, err := httpx.GetNullIntQuery(ctx, "status_code")
	if err == nil && statusCode != 0 {
		filter.StatusCode = &statusCode
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
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.EndpointHistories(count) }) {
		return
	}

	histories, total, err := c.u.Audit.History.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, histories, total, true)
}
