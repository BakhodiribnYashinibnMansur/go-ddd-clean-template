package admin

import (
	"net/http"

	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreatePermissionPost(ctx *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Name is required"})
		return
	}

	perm := &domain.Permission{
		ID:   uuid.New(),
		Name: req.Name,
	}

	err := h.uc.Authz.Permission().Create(ctx.Request.Context(), perm)
	if err != nil {
		h.l.Errorw("failed to create permission", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Permission created", "id": perm.ID})
}

func (h *Handler) DeletePermissionPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	err = h.uc.Authz.Permission().Delete(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to delete permission", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Permission deleted"})
}

func (h *Handler) CreateScopePost(ctx *gin.Context) {
	var req struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	if req.Path == "" || req.Method == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Path and Method are required"})
		return
	}

	scope := &domain.Scope{
		Path:   req.Path,
		Method: req.Method,
	}

	err := h.uc.Authz.Scope().Create(ctx.Request.Context(), scope)
	if err != nil {
		h.l.Errorw("failed to create scope", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Scope created"})
}

func (h *Handler) DeleteScopePost(ctx *gin.Context) {
	var req struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	err := h.uc.Authz.Scope().Delete(ctx.Request.Context(), req.Path, req.Method)
	if err != nil {
		h.l.Errorw("failed to delete scope", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Scope deleted"})
}
