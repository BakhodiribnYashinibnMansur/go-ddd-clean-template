package admin

import (
	"net/http"

	"gct/consts"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) ResolveSystemError(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	// Get current user ID for resolved_by
	var resolvedBy uuid.UUID
	if u, exists := ctx.Get(consts.CtxUser); exists {
		if usr, ok := u.(*domain.User); ok {
			resolvedBy = usr.ID
		}
	}

	err = h.uc.Repo.Persistent.Postgres.SystemError.MarkAsResolved(ctx.Request.Context(), id, resolvedBy)
	if err != nil {
		h.l.Errorw("failed to resolve system error", "id", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Error resolved"})
}
