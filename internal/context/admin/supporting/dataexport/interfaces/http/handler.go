package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/dataexport"
	"gct/internal/context/admin/supporting/dataexport/application/command"
	"gct/internal/context/admin/supporting/dataexport/application/query"
	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"
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

// @Summary Create a data export
// @Description Create a new data export request
// @Tags DataExports
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Data export data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /data-exports [post]
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

// @Summary List data exports
// @Description Get a paginated list of data exports
// @Tags DataExports
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /data-exports [get]
// List returns a paginated list of data exports.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListDataExportsQuery{
		Filter: exportrepo.DataExportFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListDataExports.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Exports, "total": result.Total})
}

// @Summary Get a data export
// @Description Get a single data export by ID
// @Tags DataExports
// @Accept json
// @Produce json
// @Param id path string true "Data Export ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /data-exports/{id} [get]
// Get returns a single data export by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := exportentity.ParseDataExportID(ctx.Param("id"))
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

// @Summary Update a data export
// @Description Update an existing data export by ID
// @Tags DataExports
// @Accept json
// @Produce json
// @Param id path string true "Data Export ID"
// @Param request body UpdateRequest true "Data export update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /data-exports/{id} [patch]
// Update updates a data export.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := exportentity.ParseDataExportID(ctx.Param("id"))
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

// @Summary Delete a data export
// @Description Delete a data export by ID
// @Tags DataExports
// @Accept json
// @Produce json
// @Param id path string true "Data Export ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /data-exports/{id} [delete]
// Delete deletes a data export.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := exportentity.ParseDataExportID(ctx.Param("id"))
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
