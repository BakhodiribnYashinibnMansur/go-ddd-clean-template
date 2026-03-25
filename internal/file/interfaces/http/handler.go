package http

import (
	"net/http"
	"strconv"

	"gct/internal/file"
	"gct/internal/file/application/command"
	"gct/internal/file/application/query"
	"gct/internal/file/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the File bounded context.
type Handler struct {
	bc *file.BoundedContext
	l  logger.Log
}

// NewHandler creates a new File HTTP handler.
func NewHandler(bc *file.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new file record.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.CreateFileCommand{
		Name:         req.Name,
		OriginalName: req.OriginalName,
		MimeType:     req.MimeType,
		Size:         req.Size,
		Path:         req.Path,
		URL:          req.URL,
		UploadedBy:   req.UploadedBy,
	}
	if err := h.bc.CreateFile.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of files.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListFilesQuery{
		Filter: domain.FileFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListFiles.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Files, "total": result.Total})
}

// Get returns a single file by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetFile.Handle(ctx.Request.Context(), query.GetFileQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
