package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/dataexport"
	"gct/internal/context/admin/supporting/dataexport/application/command"
	"gct/internal/context/admin/supporting/dataexport/application/query"
	"gct/internal/context/admin/supporting/dataexport/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the DataExport bounded context.
type Handler struct {
	bc *dataexport.BoundedContext
	l  logger.Log
}

// NewHandler creates a new DataExport HTTP handler.
func NewHandler(bc *dataexport.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new data export request.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateDataExportCommand{
		UserID:   req.UserID,
		DataType: req.DataType,
		Format:   req.Format,
	}
	if err := h.bc.CreateDataExport.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of data exports.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListDataExports.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Exports, "total": result.Total})
}

// Get returns a single data export by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseDataExportID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetDataExport.Handle(ctx.Request.Context(), query.GetDataExportQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a data export.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseDataExportID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateDataExportCommand{
		ID:      id,
		Status:  req.Status,
		FileURL: req.FileURL,
		Error:   req.Error,
	}
	if err := h.bc.UpdateDataExport.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a data export.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseDataExportID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteDataExport.Handle(ctx.Request.Context(), command.DeleteDataExportCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
