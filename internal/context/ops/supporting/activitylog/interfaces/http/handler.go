package http

import (
	"net/http"
	"time"

	"gct/internal/context/ops/supporting/activitylog"
	"gct/internal/context/ops/supporting/activitylog/application/query"
	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the ActivityLog bounded context.
type Handler struct {
	bc *activitylog.BoundedContext
	l  logger.Log
}

// NewHandler creates a new ActivityLog HTTP handler.
func NewHandler(bc *activitylog.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// @Summary List activity logs
// @Description Get a paginated list of field-level change history
// @Tags ActivityLog
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param entity_type query string false "Filter by entity type (e.g. user, role)"
// @Param entity_id query string false "Filter by entity ID (UUID)"
// @Param actor_id query string false "Filter by actor ID (UUID)"
// @Param field_name query string false "Filter by field name"
// @Param action query string false "Filter by action (e.g. user.updated)"
// @Param from query string false "Filter from date (RFC3339)"
// @Param to query string false "Filter to date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /activity-logs [get]
// ListActivityLogs returns a paginated list of activity log entries.
func (h *Handler) ListActivityLogs(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	filter := domain.ActivityLogFilter{
		Pagination: &pg,
	}

	if v := ctx.Query("entity_type"); v != "" {
		filter.EntityType = &v
	}
	if v := ctx.Query("entity_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.EntityID = &id
		}
	}
	if v := ctx.Query("actor_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ActorID = &id
		}
	}
	if v := ctx.Query("field_name"); v != "" {
		filter.FieldName = &v
	}
	if v := ctx.Query("action"); v != "" {
		filter.Action = &v
	}
	if v := ctx.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.FromDate = &t
		}
	}
	if v := ctx.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.ToDate = &t
		}
	}

	result, err := h.bc.ListActivityLogs.Handle(ctx.Request.Context(), query.ListActivityLogsQuery{
		Filter: filter,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result.Entries, "total": result.Total})
}
