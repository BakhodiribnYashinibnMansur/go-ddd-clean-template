package http

import (
	"net/http"

	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/content/generic/translation"
	"gct/internal/context/content/generic/translation/application/command"
	"gct/internal/context/content/generic/translation/application/query"
	"gct/internal/context/content/generic/translation/domain"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Translation bounded context.
type Handler struct {
	bc *translation.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Translation HTTP handler.
func NewHandler(bc *translation.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new translation.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateTranslationCommand{
		Key:      req.Key,
		Language: req.Language,
		Value:    req.Value,
		Group:    req.Group,
	}
	if err := h.bc.CreateTranslation.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of translations.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListTranslations.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Translations, "total": result.Total})
}

// Get returns a single translation by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseTranslationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetTranslation.Handle(ctx.Request.Context(), query.GetTranslationQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a translation.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseTranslationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateTranslationCommand{
		ID:       id,
		Key:      req.Key,
		Language: req.Language,
		Value:    req.Value,
		Group:    req.Group,
	}
	if err := h.bc.UpdateTranslation.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a translation.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseTranslationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteTranslation.Handle(ctx.Request.Context(), command.DeleteTranslationCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
