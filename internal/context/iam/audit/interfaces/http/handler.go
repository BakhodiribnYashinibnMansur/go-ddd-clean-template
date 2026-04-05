package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/iam/audit"
	"gct/internal/context/iam/audit/application/query"
	"gct/internal/context/iam/audit/domain"
	shared "gct/internal/platform/domain"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Audit bounded context.
type Handler struct {
	bc *audit.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Audit HTTP handler.
func NewHandler(bc *audit.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// ListAuditLogs returns a paginated list of audit log entries.
func (h *Handler) ListAuditLogs(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{
			Pagination: &shared.Pagination{Limit: limit, Offset: offset},
		},
	}
	result, err := h.bc.ListAuditLogs.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.AuditLogs, "total": result.Total})
}

// ListEndpointHistory returns a paginated list of endpoint history entries.
func (h *Handler) ListEndpointHistory(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{
			Pagination: &shared.Pagination{Limit: limit, Offset: offset},
		},
	}
	result, err := h.bc.ListEndpointHistory.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Entries, "total": result.Total})
}
