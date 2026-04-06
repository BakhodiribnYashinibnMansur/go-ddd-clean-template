package http

import (
	"net/http"

	"gct/internal/context/iam/supporting/audit"
	"gct/internal/context/iam/supporting/audit/application/query"
	"gct/internal/context/iam/supporting/audit/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

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

// @Summary List audit logs
// @Description Get a paginated list of audit log entries
// @Tags Audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /audit-logs [get]
// ListAuditLogs returns a paginated list of audit log entries.
func (h *Handler) ListAuditLogs(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{
			Pagination: &pg,
		},
	}
	result, err := h.bc.ListAuditLogs.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.AuditLogs, "total": result.Total})
}

// @Summary List endpoint history
// @Description Get a paginated list of endpoint history entries
// @Tags Audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /endpoint-history [get]
// ListEndpointHistory returns a paginated list of endpoint history entries.
func (h *Handler) ListEndpointHistory(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{
			Pagination: &pg,
		},
	}
	result, err := h.bc.ListEndpointHistory.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Entries, "total": result.Total})
}
