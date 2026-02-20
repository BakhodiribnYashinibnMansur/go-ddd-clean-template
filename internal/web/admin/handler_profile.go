package admin

import (
	"fmt"
	"gct/consts"
	"gct/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfilePost handles users updating their own profile
func (h *Handler) ProfilePost(ctx *gin.Context) {
	// 1. Get current session from context
	sessVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		ctx.Redirect(http.StatusFound, "/admin/login")
		return
	}
	sess, ok := sessVal.(*domain.Session)
	if !ok || sess == nil {
		h.l.Errorw("ProfilePost - invalid session in context")
		ctx.Redirect(http.StatusFound, "/admin/login")
		return
	}

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
