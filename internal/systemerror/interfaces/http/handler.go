package http

import (
	"net/http"
	"strconv"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/systemerror"
	"gct/internal/systemerror/application/command"
	"gct/internal/systemerror/application/query"
	"gct/internal/systemerror/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of system errors.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListSystemErrorsQuery{
		Filter: domain.SystemErrorFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListSystemErrors.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Errors, "total": result.Total})
}

// Get returns a single system error by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetSystemError.Handle(ctx.Request.Context(), query.GetSystemErrorQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Resolve marks a system error as resolved.
func (h *Handler) Resolve(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req ResolveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.ResolveErrorCommand{
		ID:         id,
		ResolvedBy: req.ResolvedBy,
	}
	if err := h.bc.ResolveError.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
