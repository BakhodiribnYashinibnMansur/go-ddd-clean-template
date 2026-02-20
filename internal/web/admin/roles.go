package admin

import (
	"net/http"

	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) RoleDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/rbac/roles")
		return
	}

	role, err := h.uc.Authz.Role().Get(ctx.Request.Context(), &domain.RoleFilter{ID: &id})
	if err != nil {
		h.l.Errorw("failed to fetch role", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/rbac/roles")
		return
	}

	// Get all permissions
	allPerms, _, _ := h.uc.Authz.Permission().Gets(ctx.Request.Context(), &domain.PermissionsFilter{
		Pagination: &domain.Pagination{Limit: 1000},
	})

	// Get role's assigned permissions
	rolePerms, _ := h.uc.Repo.Persistent.Postgres.Authz.Role.GetPermissions(ctx.Request.Context(), id)
	assignedMap := make(map[uuid.UUID]bool)
	for _, p := range rolePerms {
		assignedMap[p.ID] = true
	}

	// Get user count for this role
	limit := &domain.Pagination{Limit: 1}
	_, usersCount, _ := h.uc.User.Client().Gets(ctx.Request.Context(), &domain.UsersFilter{
		Pagination: limit,
		UserFilter: domain.UserFilter{RoleID: &id},
	})

	h.servePage(ctx, "rbac/role_detail.html", "Role: "+role.Name, "roles", map[string]any{
		"Role":            role,
		"AllPermissions":  allPerms,
		"RolePermissions": rolePerms,
		"AssignedMap":     assignedMap,
		"UsersCount":      usersCount,
	})
}

func (h *Handler) CreateRolePost(ctx *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	role := &domain.Role{
		ID:   uuid.New(),
		Name: req.Name,
	}

	err := h.uc.Authz.Role().Create(ctx.Request.Context(), role)
	if err != nil {
		h.l.Errorw("failed to create role", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Role created", "id": role.ID})
}

func (h *Handler) UpdateRolePost(ctx *gin.Context) {
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

	role, err := h.uc.Authz.Role().Get(ctx.Request.Context(), &domain.RoleFilter{ID: &id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Role not found"})
		return
	}

	role.Name = req.Name
	err = h.uc.Authz.Role().Update(ctx.Request.Context(), role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Role updated"})
}

func (h *Handler) DeleteRolePost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	err = h.uc.Authz.Role().Delete(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to delete role", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Role deleted"})
}

func (h *Handler) RoleAddPermission(ctx *gin.Context) {
	roleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid role ID"})
		return
	}
	permID, err := uuid.Parse(ctx.Param("pid"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid permission ID"})
		return
	}

	err = h.uc.Authz.Role().AddPermission(ctx.Request.Context(), roleID, permID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Permission added"})
}

func (h *Handler) RoleRemovePermission(ctx *gin.Context) {
	roleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid role ID"})
		return
	}
	permID, err := uuid.Parse(ctx.Param("pid"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid permission ID"})
		return
	}

	err = h.uc.Authz.Role().RemovePermission(ctx.Request.Context(), roleID, permID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Permission removed"})
}
