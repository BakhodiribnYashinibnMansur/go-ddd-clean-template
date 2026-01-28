package admin

import (
	"fmt"
	"gct/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfilePost handles users updating their own profile
func (h *Handler) ProfilePost(ctx *gin.Context) {
	// 1. Get current user from context/session
	sessVal, _ := ctx.Get("session")
	sess := sessVal.(*domain.Session) // safe assumption if middleware works

	// 2. Bind Form Data
	var form struct {
		Username string `form:"username"`
		Phone    string `form:"phone"`
		Password string `form:"password"`
	}

	if err := ctx.ShouldBind(&form); err != nil {
		h.l.Errorw("ProfilePost - Bind error", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/profile?error=Invalid+form+data")
		return
	}

	// 3. Prepare Update
	update := &domain.User{
		ID: sess.UserID,
	}

	// Only update fields if provided
	if form.Username != "" {
		update.Username = &form.Username
	}
	if form.Phone != "" {
		update.Phone = &form.Phone
	}
	// Password is treated specially in Update usecase logic (if != "", hashes and sets)
	if form.Password != "" {
		update.Password = form.Password
	}

	// 4. Call Usecase (User.Client.Update)
	err := h.uc.User.Client().Update(ctx.Request.Context(), update)
	if err != nil {
		h.l.Errorw("ProfilePost - Update error", "error", err)
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/admin/profile?error=Update+failed:+%v", err))
		return
	}

	// 5. Success
	ctx.Redirect(http.StatusFound, "/admin/profile?success=Profile+updated+successfully")
}
