package admin

import (
	"net/http"

	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) Relations(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	pagination.Limit = 100 // Show all relations
	filter := &domain.RelationsFilter{Pagination: pagination}

	if t := ctx.Query("type"); t != "" {
		filter.Type = &t
	}
	if name := ctx.Query("name"); name != "" {
		filter.Name = &name
	}

	relations, count, err := h.uc.Authz.Relation().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch relations", "error", err)
	}

	// Group by type
	grouped := map[string][]*domain.Relation{
		"REGION":     {},
		"BRANCH":     {},
		"UNREVEALED": {},
	}
	for _, r := range relations {
		grouped[string(r.Type)] = append(grouped[string(r.Type)], r)
	}

	h.servePage(ctx, "relations.html", "Relations", "relations", map[string]any{
		"Relations":   relations,
		"Grouped":     grouped,
		"TotalCount":  count,
		"FilterType":  ctx.Query("type"),
		"FilterName":  ctx.Query("name"),
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) CreateRelationPost(ctx *gin.Context) {
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	relation := &domain.Relation{
		ID:   uuid.New(),
		Name: req.Name,
		Type: domain.RelationType(req.Type),
	}

	err := h.uc.Authz.Relation().Create(ctx.Request.Context(), relation)
	if err != nil {
		h.l.Errorw("failed to create relation", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Relation created"})
}

func (h *Handler) UpdateRelationPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	existing, err := h.uc.Authz.Relation().Get(ctx.Request.Context(), &domain.RelationFilter{ID: &id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Relation not found"})
		return
	}

	existing.Name = req.Name
	err = h.uc.Authz.Relation().Update(ctx.Request.Context(), existing)
	if err != nil {
		h.l.Errorw("failed to update relation", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Relation updated"})
}

func (h *Handler) DeleteRelationPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	err = h.uc.Authz.Relation().Delete(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to delete relation", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Relation deleted"})
}
