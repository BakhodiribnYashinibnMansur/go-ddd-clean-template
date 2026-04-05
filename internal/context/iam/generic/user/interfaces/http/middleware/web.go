package middleware

import (
	"net/http"
	"net/url"

	"gct/internal/context/iam/generic/user/application/query"
	userdomain "gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

// AuthWeb wraps authentication for web/admin panel, supporting interactive
// browser-based flows.
//
// This middleware validates the access token and, on success, injects session
// and user data into the Gin context. If the token is expired or invalid,
// it redirects to the login page with a return URL for seamless user experience.
//
// Note: Silent token rotation (auto-refresh) is not yet implemented in the DDD
// version. When the access token expires, the user is redirected to login.
func (m *AuthMiddleware) AuthWeb(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		// Fallback to login redirect for web interaction
		ctx.Redirect(http.StatusFound, "/admin/login?return_url="+url.QueryEscape(ctx.Request.RequestURI))
		ctx.Abort()
		return
	}

	// Standard identity injection
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	// Fetch user for template context (role title, etc.)
	user, uErr := m.findUserForAuth.Handle(ctx.Request.Context(), query.FindUserForAuthQuery{UserID: userdomain.UserID(session.UserID)})
	if uErr == nil {
		ctx.Set(consts.CtxUser, user)
	} else {
		m.l.Warnw("AuthWeb - Failed to fetch user for context", "user_id", session.UserID, "error", uErr)
	}

	ctx.Next()
}
