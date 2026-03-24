package auth

import (
	"errors"
	"net/http"
	"net/url"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/security/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthWeb wraps authentication for web/admin panel, supporting interactive features like auto-refreshing expired tokens.
//
// This middleware is specifically designed for browser-based clients and implements
// automatic token refresh when the access token expires. If refresh fails, it redirects
// to the login page with a return URL for seamless user experience.
//
// Key features:
// - Automatic silent token rotation on expiry
// - Graceful fallback to login redirect
// - User and role context injection for templates
func (m *AuthMiddleware) AuthWeb(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		// Attempt logical auto-refresh if token is simply expired but refresh token exists
		if errors.Is(err, httpx.ErrExpiredToken) {
			refreshToken := cookie.GetCookie(ctx, consts.CookieRefreshToken)
			if refreshToken != "" {
				rt, pErr := jwt.ParseRefreshToken(refreshToken)

				if pErr == nil {
					sID, _ := uuid.Parse(rt.SessionID)
					sess, sErr := m.sessionuc.Get(ctx, &domain.SessionFilter{ID: &sID})

					if sErr == nil && !sess.Revoked && !sess.IsExpired() && rt.Verify(sess.RefreshTokenHash) {
						// Perform silent Token Rotation
						res, rErr := m.userUC.RotateSession(ctx.Request.Context(), &domain.RefreshIn{SessionID: sID})
						if rErr == nil {
							isSecure := ctx.Request.TLS != nil || ctx.Request.Header.Get("X-Forwarded-Proto") == "https"

							ctx.SetCookie(consts.CookieAccessToken, res.AccessToken, int(m.cfg.JWT.AccessTTL.Seconds()), "/", "", isSecure, true)
							ctx.SetCookie(consts.CookieRefreshToken, res.RefreshToken, int(m.cfg.JWT.RefreshTTL.Seconds()), "/", "", isSecure, true)

							freshSess, fErr := m.sessionuc.Get(ctx, &domain.SessionFilter{ID: &sID})
							if fErr != nil {
								m.l.Errorw("AuthWeb - Failed to fetch fresh session", "error", fErr)
								ctx.Redirect(http.StatusFound, "/admin/login")
								ctx.Abort()
								return
							}

							ctx.Set(consts.CtxSessionID, freshSess.ID)
							ctx.Set(consts.CtxSession, freshSess)
							ctx.Set(consts.CtxUserID, freshSess.UserID.String())

							u, uErr := m.userUC.Get(ctx, &domain.UserFilter{ID: &freshSess.UserID})
							if uErr == nil {
								ctx.Set(consts.CtxUser, u)
							}
							ctx.Next()
							return
						} else {
							m.l.Warnw("AuthWeb - Auto-refresh rotation failed", "error", rErr)
						}
					}
				}
			}
		}

		// Fallback to login redirect for web interaction
		ctx.Redirect(http.StatusFound, "/admin/login?return_url="+url.QueryEscape(ctx.Request.RequestURI))
		ctx.Abort()
		return
	}

	// standard identity injection
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	user, uErr := m.userUC.Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if uErr == nil {
		ctx.Set(consts.CtxUser, user)

		if user.RoleID != nil {
			role, rErr := m.authzUC.Role().Get(ctx, &domain.RoleFilter{ID: user.RoleID})
			if rErr == nil {
				ctx.Set(consts.CtxRoleTitle, role.Name)
			} else {
				m.l.Warnw("AuthWeb - Failed to fetch role", "role_id", user.RoleID, "error", rErr)
			}
		}
	} else {
		m.l.Warnw("AuthWeb - Failed to fetch user for context", "user_id", session.UserID, "error", uErr)
	}

	ctx.Next()
}
