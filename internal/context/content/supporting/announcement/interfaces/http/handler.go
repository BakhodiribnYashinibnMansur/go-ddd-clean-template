package http

import (
	"net/http"

	"gct/internal/context/content/supporting/announcement"
	"gct/internal/context/content/supporting/announcement/application/command"
	"gct/internal/context/content/supporting/announcement/application/query"
	"gct/internal/context/content/supporting/announcement/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Announcement bounded context.
type Handler struct {
	bc *announcement.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Announcement HTTP handler.
func NewHandler(bc *announcement.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// @Summary Create an announcement
// @Description Create a new announcement
// @Tags Announcements
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Announcement data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /announcements [post]
// Create creates a new announcement.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateAnnouncementCommand{
		Title:     req.Title,
		Content:   req.Content,
		Priority:  req.Priority,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	if err := h.bc.CreateAnnouncement.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List announcements
// @Description Get a paginated list of announcements
// @Tags Announcements
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /announcements [get]
// List returns a paginated list of announcements.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListAnnouncements.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Announcements, "total": result.Total})
}

// @Summary Get an announcement
// @Description Get an announcement by ID
// @Tags Announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /announcements/{id} [get]
// Get returns a single announcement by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseAnnouncementID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetAnnouncement.Handle(ctx.Request.Context(), query.GetAnnouncementQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Update an announcement
// @Description Update an existing announcement by ID
// @Tags Announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param request body UpdateRequest true "Announcement data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /announcements/{id} [patch]
// Update updates an announcement.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseAnnouncementID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateAnnouncementCommand{
		ID:        id,
		Title:     req.Title,
		Content:   req.Content,
		Priority:  req.Priority,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Publish:   req.Publish,
	}
	if err := h.bc.UpdateAnnouncement.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete an announcement
// @Description Delete an announcement by ID
// @Tags Announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /announcements/{id} [delete]
// Delete deletes an announcement.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseAnnouncementID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteAnnouncement.Handle(ctx.Request.Context(), command.DeleteAnnouncementCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
