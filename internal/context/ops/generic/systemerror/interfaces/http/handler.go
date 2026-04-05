package http

import (
	"net/http"

	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/ops/generic/systemerror"
	"gct/internal/context/ops/generic/systemerror/application/command"
	"gct/internal/context/ops/generic/systemerror/application/query"
	"gct/internal/context/ops/generic/systemerror/domain"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the SystemError bounded context.
type Handler struct {
	bc *systemerror.BoundedContext
	l  logger.Log
}

// NewHandler creates a new SystemError HTTP handler.
func NewHandler(bc *systemerror.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create records a new system error.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateSystemErrorCommand{
		Code:        req.Code,
		Message:     req.Message,
		StackTrace:  req.StackTrace,
		Metadata:    req.Metadata,
		Severity:    req.Severity,
		ServiceName: req.ServiceName,
		RequestID:   req.RequestID,
		UserID:      req.UserID,
		IPAddress:   req.IPAddress,
		Path:        req.Path,
		Method:      req.Method,
	}
	if err := h.bc.CreateSystemError.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of system errors.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListSystemErrorsQuery{
		Filter: domain.SystemErrorFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListSystemErrors.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Errors, "total": result.Total})
}

// Get returns a single system error by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseSystemErrorID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetSystemError.Handle(ctx.Request.Context(), query.GetSystemErrorQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Resolve marks a system error as resolved.
func (h *Handler) Resolve(ctx *gin.Context) {
	id, err := domain.ParseSystemErrorID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req ResolveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.ResolveErrorCommand{
		ID:         id,
		ResolvedBy: req.ResolvedBy,
	}
	if err := h.bc.ResolveError.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
