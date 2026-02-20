package admin

import (
	"net/http"

	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) PolicyDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/abac/policies")
		return
	}

	policy, err := h.uc.Authz.Policy().Get(ctx.Request.Context(), &domain.PolicyFilter{ID: &id})
	if err != nil {
		h.l.Errorw("failed to fetch policy", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/abac/policies")
		return
	}

	// Get all permissions for the dropdown
	allPerms, _, _ := h.uc.Authz.Permission().Gets(ctx.Request.Context(), &domain.PermissionsFilter{
		Pagination: &domain.Pagination{Limit: 1000},
	})

	// Get permission name for this policy
	var permName string
	perm, err := h.uc.Authz.Permission().Get(ctx.Request.Context(), &domain.PermissionFilter{ID: &policy.PermissionID})
	if err == nil {
		permName = perm.Name
	}

	h.servePage(ctx, "abac/policy_detail.html", "Policy Detail", "policies", map[string]any{
		"Policy":         policy,
		"AllPermissions": allPerms,
		"PermissionName": permName,
	})
}

func (h *Handler) CreatePolicyPage(ctx *gin.Context) {
	allPerms, _, _ := h.uc.Authz.Permission().Gets(ctx.Request.Context(), &domain.PermissionsFilter{
		Pagination: &domain.Pagination{Limit: 1000},
	})

	h.servePage(ctx, "abac/policy_detail.html", "Create Policy", "policies", map[string]any{
		"Policy":         nil,
		"AllPermissions": allPerms,
		"IsNew":          true,
	})
}

func (h *Handler) CreatePolicyPost(ctx *gin.Context) {
	var req struct {
		PermissionID string         `json:"permission_id"`
		Effect       string         `json:"effect"`
		Priority     int            `json:"priority"`
		Active       bool           `json:"active"`
		Conditions   map[string]any `json:"conditions"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid permission ID"})
		return
	}

	policy := &domain.Policy{
		ID:           uuid.New(),
		PermissionID: permID,
		Effect:       domain.PolicyEffect(req.Effect),
		Priority:     req.Priority,
		Active:       req.Active,
		Conditions:   req.Conditions,
	}

	if policy.Conditions == nil {
		policy.Conditions = make(map[string]any)
	}

	err = h.uc.Authz.Policy().Create(ctx.Request.Context(), policy)
	if err != nil {
		h.l.Errorw("failed to create policy", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Policy created", "id": policy.ID})
}

func (h *Handler) UpdatePolicyPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	var req struct {
		PermissionID string         `json:"permission_id"`
		Effect       string         `json:"effect"`
		Priority     int            `json:"priority"`
		Active       bool           `json:"active"`
		Conditions   map[string]any `json:"conditions"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid permission ID"})
		return
	}

	policy := &domain.Policy{
		ID:           id,
		PermissionID: permID,
		Effect:       domain.PolicyEffect(req.Effect),
		Priority:     req.Priority,
		Active:       req.Active,
		Conditions:   req.Conditions,
	}

	if policy.Conditions == nil {
		policy.Conditions = make(map[string]any)
	}

	err = h.uc.Authz.Policy().Update(ctx.Request.Context(), policy)
	if err != nil {
		h.l.Errorw("failed to update policy", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Policy updated"})
}

func (h *Handler) DeletePolicyPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	err = h.uc.Authz.Policy().Delete(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to delete policy", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Policy deleted"})
}

func (h *Handler) TogglePolicyActive(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	policy, err := h.uc.Authz.Policy().Get(ctx.Request.Context(), &domain.PolicyFilter{ID: &id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Policy not found"})
		return
	}

	policy.Active = !policy.Active
	err = h.uc.Authz.Policy().Update(ctx.Request.Context(), policy)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Policy toggled", "active": policy.Active})
}
