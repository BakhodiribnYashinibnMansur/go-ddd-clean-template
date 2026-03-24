package log

import (
	"net/http"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gets godoc
// @Summary     Get audit logs
// @Description Retrieve audit logs with filtering and pagination
// @Tags        audit
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       user_id query string false "User ID"
// @Param       action query string false "Action"
// @Param       resource_type query string false "Resource Type"
// @Param       resource_id query string false "Resource ID"
// @Param       success query bool false "Success"
// @Param       from_date query string false "From Date (RFC3339)"
// @Param       to_date query string false "To Date (RFC3339)"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /audit/logs [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - audit - log - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.AuditLogsFilter{
		Pagination: &pagination,
		AuditLogFilter: domain.AuditLogFilter{
			ResourceType: func() *string {
				s := httpx.GetNullStringQuery(ctx, consts.QueryResourceType)
				if s == "" {
					return nil
				}
				return &s
			}(),
		},
	}

	userIDStr := httpx.GetNullStringQuery(ctx, consts.QueryUserID)
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &uid
		}
	}

	actionStr := httpx.GetNullStringQuery(ctx, consts.QueryAction)
	if actionStr != "" {
		act := domain.AuditActionType(actionStr)
		filter.Action = &act
	}

	reqIDStr := httpx.GetNullStringQuery(ctx, consts.QueryResourceID)
	if reqIDStr != "" {
		rid, err := uuid.Parse(reqIDStr)
		if err == nil {
			filter.ResourceID = &rid
		}
	}

	successStr := httpx.GetNullStringQuery(ctx, consts.QuerySuccess)
	switch successStr {
	case "true":
		t := true
		filter.Success = &t
	case "false":
		f := false
		filter.Success = &f
	}

	fromDateStr := httpx.GetNullStringQuery(ctx, consts.QueryFromDate)
	if fromDateStr != "" {
		t, err := time.Parse(time.RFC3339, fromDateStr)
		if err == nil {
			filter.FromDate = &t
		}
	}

	toDateStr := httpx.GetNullStringQuery(ctx, consts.QueryToDate)
	if toDateStr != "" {
		t, err := time.Parse(time.RFC3339, toDateStr)
		if err == nil {
			filter.ToDate = &t
		}
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.AuditLogs(count) }) {
		return
	}

	logs, total, err := c.u.Audit.Log().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, logs, total, true)
}
