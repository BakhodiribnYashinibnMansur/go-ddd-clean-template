package app

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user"
	"gct/internal/user/application/command"

	"github.com/google/uuid"
	miniogo "github.com/minio/minio-go/v7"

	announcementhttp "gct/internal/announcement/interfaces/http"
	audithttp "gct/internal/audit/interfaces/http"
	authzhttp "gct/internal/authz/interfaces/http"
	dashboardhttp "gct/internal/dashboard/interfaces/http"
	dataexporthttp "gct/internal/dataexport/interfaces/http"
	errorcodehttp "gct/internal/errorcode/interfaces/http"
	featureflaghttp "gct/internal/featureflag/interfaces/http"
	filehttp "gct/internal/file/interfaces/http"
	integrationhttp "gct/internal/integration/interfaces/http"
	iprulehttp "gct/internal/iprule/interfaces/http"

	metrichttp "gct/internal/metric/interfaces/http"
	notificationhttp "gct/internal/notification/interfaces/http"
	ratelimithttp "gct/internal/ratelimit/interfaces/http"
	sessionhttp "gct/internal/session/interfaces/http"
	sitesettinghttp "gct/internal/sitesetting/interfaces/http"
	systemerrorhttp "gct/internal/systemerror/interfaces/http"
	translationhttp "gct/internal/translation/interfaces/http"
	userhttp "gct/internal/user/interfaces/http"
	usersettinghttp "gct/internal/usersetting/interfaces/http"


	"github.com/gin-gonic/gin"
)

// sessionRevokerAdapter bridges the session HTTP handler to the User BC's sign-out commands.
type sessionRevokerAdapter struct {
	userBC *user.BoundedContext
}

func (a *sessionRevokerAdapter) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return a.userBC.SignOut.Handle(ctx, command.SignOutCommand{
		UserID:    userID,
		SessionID: sessionID,
	})
}

func (a *sessionRevokerAdapter) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return a.userBC.RevokeAll.Handle(ctx, command.RevokeAllSessionsCommand{
		UserID: userID,
	})
}

// RouteOptions holds optional dependencies for route registration.
type RouteOptions struct {
	Minio       *miniogo.Client
	MinioBucket string
}

// RegisterDDDRoutes registers all HTTP routes for DDD bounded contexts.
func RegisterDDDRoutes(router *gin.Engine, bcs *DDDBoundedContexts, authMW, authzMW, csrfMW gin.HandlerFunc, l logger.Log, opts ...RouteOptions) {
	var opt RouteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	userHandler := userhttp.NewHandler(bcs.User, l)
	authzHandler := authzhttp.NewHandler(bcs.Authz, l)
	sessionHandler := sessionhttp.NewHandler(bcs.Session, l)
	auditHandler := audithttp.NewHandler(bcs.Audit, l)
	dashboardHandler := dashboardhttp.NewHandler(bcs.Dashboard, l)
	sysErrHandler := systemerrorhttp.NewHandler(bcs.SystemError, l)
	metricHandler := metrichttp.NewHandler(bcs.Metric, l)
	ffHandler := featureflaghttp.NewHandler(bcs.FeatureFlag, l)
	integrationHandler := integrationhttp.NewHandler(bcs.Integration, l)

	notifHandler := notificationhttp.NewHandler(bcs.Notification, l)
	announcementHandler := announcementhttp.NewHandler(bcs.Announcement, l)
	translationHandler := translationhttp.NewHandler(bcs.Translation, l)
	siteSettingHandler := sitesettinghttp.NewHandler(bcs.SiteSetting, l)
	rateLimitHandler := ratelimithttp.NewHandler(bcs.RateLimit, l)
	ipRuleHandler := iprulehttp.NewHandler(bcs.IPRule, l)

	dataExportHandler := dataexporthttp.NewHandler(bcs.DataExport, l)
	fileHandler := filehttp.NewHandler(bcs.File, l)
	if opt.Minio != nil {
		fileHandler.SetMinio(opt.Minio, opt.MinioBucket)
	}
	userSettingHandler := usersettinghttp.NewHandler(bcs.UserSetting, l)
	errorCodeHandler := errorcodehttp.NewHandler(bcs.ErrorCode, l)

	// Public routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/sign-in", userHandler.SignIn)
		auth.POST("/sign-up", userHandler.SignUp)
	}

	// Auth-only
	authOnly := router.Group("/api/v1/auth")
	authOnly.Use(authMW)
	{
		authOnly.POST("/sign-out", userHandler.SignOut)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(authMW, authzMW, csrfMW)

	users := protected.Group("/users")
	{
		users.POST("", userHandler.Create)
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.PATCH("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
		users.POST("/:id/approve", userHandler.Approve)
		users.POST("/:id/role", userHandler.ChangeRole)
		users.POST("/bulk-action", userHandler.BulkAction)
	}

	roles := protected.Group("/roles")
	{
		roles.POST("", authzHandler.CreateRole)
		roles.GET("", authzHandler.ListRoles)
		roles.GET("/:id", authzHandler.GetRole)
		roles.PATCH("/:id", authzHandler.UpdateRole)
		roles.DELETE("/:id", authzHandler.DeleteRole)
		roles.POST("/:id/permissions", authzHandler.AssignPermission)
	}

	permissions := protected.Group("/permissions")
	{
		permissions.POST("", authzHandler.CreatePermission)
		permissions.GET("", authzHandler.ListPermissions)
		permissions.DELETE("/:id", authzHandler.DeletePermission)
		permissions.POST("/:id/scopes", authzHandler.AssignScope)
	}

	policies := protected.Group("/policies")
	{
		policies.POST("", authzHandler.CreatePolicy)
		policies.GET("", authzHandler.ListPolicies)
		policies.PATCH("/:id", authzHandler.UpdatePolicy)
		policies.DELETE("/:id", authzHandler.DeletePolicy)
		policies.POST("/:id/toggle", authzHandler.TogglePolicy)
	}

	scopes := protected.Group("/scopes")
	{
		scopes.POST("", authzHandler.CreateScope)
		scopes.GET("", authzHandler.ListScopes)
		scopes.DELETE("", authzHandler.DeleteScope)
	}

	sessionHandler.SetRevoker(&sessionRevokerAdapter{userBC: bcs.User})

	sessions := protected.Group("/sessions")
	{
		sessions.GET("", sessionHandler.List)
		sessions.GET("/:id", sessionHandler.Get)
		sessions.DELETE("/:id", sessionHandler.Delete)
		sessions.POST("/revoke-all", sessionHandler.RevokeAll)
	}

	auditLogs := protected.Group("/audit-logs")
	{
		auditLogs.GET("", auditHandler.ListAuditLogs)
	}

	endpointHistory := protected.Group("/endpoint-history")
	{
		endpointHistory.GET("", auditHandler.ListEndpointHistory)
	}

	dashboard := protected.Group("/dashboard")
	{
		dashboard.GET("/stats", dashboardHandler.GetStats)
	}

	systemErrors := protected.Group("/system-errors")
	{
		systemErrors.POST("", sysErrHandler.Create)
		systemErrors.GET("", sysErrHandler.List)
		systemErrors.GET("/:id", sysErrHandler.Get)
		systemErrors.POST("/:id/resolve", sysErrHandler.Resolve)
	}

	metrics := protected.Group("/metrics")
	{
		metrics.POST("", metricHandler.Create)
		metrics.GET("", metricHandler.List)
	}

	ffHandler.RegisterRoutes(protected)

	integrations := protected.Group("/integrations")
	{
		integrations.POST("", integrationHandler.Create)
		integrations.GET("", integrationHandler.List)
		integrations.GET("/:id", integrationHandler.Get)
		integrations.PATCH("/:id", integrationHandler.Update)
		integrations.DELETE("/:id", integrationHandler.Delete)
	}


	notifications := protected.Group("/notifications")
	{
		notifications.POST("", notifHandler.Create)
		notifications.GET("", notifHandler.List)
		notifications.GET("/:id", notifHandler.Get)
		notifications.DELETE("/:id", notifHandler.Delete)
	}

	announcements := protected.Group("/announcements")
	{
		announcements.POST("", announcementHandler.Create)
		announcements.GET("", announcementHandler.List)
		announcements.GET("/:id", announcementHandler.Get)
		announcements.PATCH("/:id", announcementHandler.Update)
		announcements.DELETE("/:id", announcementHandler.Delete)
	}

	translations := protected.Group("/translations")
	{
		translations.POST("", translationHandler.Create)
		translations.GET("", translationHandler.List)
		translations.GET("/:id", translationHandler.Get)
		translations.PATCH("/:id", translationHandler.Update)
		translations.DELETE("/:id", translationHandler.Delete)
	}

	siteSettings := protected.Group("/site-settings")
	{
		siteSettings.POST("", siteSettingHandler.Create)
		siteSettings.GET("", siteSettingHandler.List)
		siteSettings.GET("/:id", siteSettingHandler.Get)
		siteSettings.PATCH("/:id", siteSettingHandler.Update)
		siteSettings.DELETE("/:id", siteSettingHandler.Delete)
	}

	rateLimits := protected.Group("/rate-limits")
	{
		rateLimits.POST("", rateLimitHandler.Create)
		rateLimits.GET("", rateLimitHandler.List)
		rateLimits.GET("/:id", rateLimitHandler.Get)
		rateLimits.PATCH("/:id", rateLimitHandler.Update)
		rateLimits.DELETE("/:id", rateLimitHandler.Delete)
	}

	ipRules := protected.Group("/ip-rules")
	{
		ipRules.POST("", ipRuleHandler.Create)
		ipRules.GET("", ipRuleHandler.List)
		ipRules.GET("/:id", ipRuleHandler.Get)
		ipRules.PATCH("/:id", ipRuleHandler.Update)
		ipRules.DELETE("/:id", ipRuleHandler.Delete)
	}


	dataExports := protected.Group("/data-exports")
	{
		dataExports.POST("", dataExportHandler.Create)
		dataExports.GET("", dataExportHandler.List)
		dataExports.GET("/:id", dataExportHandler.Get)
		dataExports.PATCH("/:id", dataExportHandler.Update)
		dataExports.DELETE("/:id", dataExportHandler.Delete)
	}

	files := protected.Group("/files")
	{
		files.POST("", fileHandler.Create)
		files.GET("", fileHandler.List)
		files.GET("/:id", fileHandler.Get)
		files.POST("/upload/image", fileHandler.UploadImage)
		files.POST("/upload/images", fileHandler.UploadImages)
		files.POST("/upload/doc", fileHandler.UploadDoc)
		files.GET("/download", fileHandler.Download)
	}

	userSettings := protected.Group("/user-settings")
	{
		userSettings.POST("", userSettingHandler.Upsert)
		userSettings.GET("", userSettingHandler.List)
		userSettings.DELETE("/:id", userSettingHandler.Delete)
	}

	errorCodes := protected.Group("/error-codes")
	{
		errorCodes.POST("", errorCodeHandler.Create)
		errorCodes.GET("", errorCodeHandler.List)
		errorCodes.GET("/:id", errorCodeHandler.Get)
		errorCodes.PATCH("/:id", errorCodeHandler.Update)
		errorCodes.DELETE("/:id", errorCodeHandler.Delete)
	}
}
