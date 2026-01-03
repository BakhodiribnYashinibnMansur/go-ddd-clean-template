package log

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
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
// @Success     200 {object} response.SuccessResponse{data=[]domain.AuditLog}
// @Failure     400 {object} response.ErrorResponse
// @Router      /audit/logs [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := util.GetPagination(ctx)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - audit - log - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.AuditLogsFilter{
		Pagination: &pagination,
		AuditLogFilter: domain.AuditLogFilter{
			ResourceType: func() *string {
				s := util.GetNullStringQuery(ctx, "resource_type")
				if s == "" {
					return nil
				}
				return &s
			}(),
		},
	}

	userIDStr := util.GetNullStringQuery(ctx, "user_id")
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &uid
		}
	}

	actionStr := util.GetNullStringQuery(ctx, "action")
	if actionStr != "" {
		act := domain.AuditActionType(actionStr)
		filter.Action = &act
	}

	reqIDStr := util.GetNullStringQuery(ctx, "resource_id")
	if reqIDStr != "" {
		rid, err := uuid.Parse(reqIDStr)
		if err == nil {
			filter.ResourceID = &rid
		}
	}

	successStr := util.GetNullStringQuery(ctx, "success")
	if successStr == "true" {
		t := true
		filter.Success = &t
	} else if successStr == "false" {
		f := false
		filter.Success = &f
	}

	fromDateStr := util.GetNullStringQuery(ctx, "from_date")
	if fromDateStr != "" {
		t, err := time.Parse(time.RFC3339, fromDateStr)
		if err == nil {
			filter.FromDate = &t
		}
	}

	toDateStr := util.GetNullStringQuery(ctx, "to_date")
	if toDateStr != "" {
		t, err := time.Parse(time.RFC3339, toDateStr)
		if err == nil {
			filter.ToDate = &t
		}
	}

	logs, total, err := c.u.Audit.Log.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, logs, total, true)
}
