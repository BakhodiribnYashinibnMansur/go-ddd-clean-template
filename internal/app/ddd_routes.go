package app

import (
	"gct/internal/shared/infrastructure/logger"

	announcementhttp "gct/internal/announcement/interfaces/http"
	audithttp "gct/internal/audit/interfaces/http"
	authzhttp "gct/internal/authz/interfaces/http"
	dashboardhttp "gct/internal/dashboard/interfaces/http"
	dataexporthttp "gct/internal/dataexport/interfaces/http"
	emailtemplatehttp "gct/internal/emailtemplate/interfaces/http"
	errorcodehttp "gct/internal/errorcode/interfaces/http"
	featureflaghttp "gct/internal/featureflag/interfaces/http"
	filehttp "gct/internal/file/interfaces/http"
	integrationhttp "gct/internal/integration/interfaces/http"
	iprulehttp "gct/internal/iprule/interfaces/http"
	jobhttp "gct/internal/job/interfaces/http"
	metrichttp "gct/internal/metric/interfaces/http"
	notificationhttp "gct/internal/notification/interfaces/http"
	ratelimithttp "gct/internal/ratelimit/interfaces/http"
	sessionhttp "gct/internal/session/interfaces/http"
	sitesettinghttp "gct/internal/sitesetting/interfaces/http"
	systemerrorhttp "gct/internal/systemerror/interfaces/http"
	translationhttp "gct/internal/translation/interfaces/http"
	userhttp "gct/internal/user/interfaces/http"
	usersettinghttp "gct/internal/usersetting/interfaces/http"
	webhookhttp "gct/internal/webhook/interfaces/http"

	"github.com/gin-gonic/gin"
)

// RegisterDDDRoutes registers HTTP routes for all DDD bounded contexts.
func RegisterDDDRoutes(router *gin.Engine, bcs *DDDBoundedContexts, authMW, authzMW, csrfMW gin.HandlerFunc, l logger.Log) {
	// Create handlers
	userHandler := userhttp.NewHandler(bcs.User, l)
	authzHandler := authzhttp.NewHandler(bcs.Authz, l)
	sessionHandler := sessionhttp.NewHandler(bcs.Session, l)
	auditHandler := audithttp.NewHandler(bcs.Audit, l)
	dashboardHandler := dashboardhttp.NewHandler(bcs.Dashboard, l)
	sysErrHandler := systemerrorhttp.NewHandler(bcs.SystemError, l)
	metricHandler := metrichttp.NewHandler(bcs.Metric, l)
	ffHandler := featureflaghttp.NewHandler(bcs.FeatureFlag, l)
	integrationHandler := integrationhttp.NewHandler(bcs.Integration, l)
	webhookHandler := webhookhttp.NewHandler(bcs.Webhook, l)
	notifHandler := notificationhttp.NewHandler(bcs.Notification, l)
	emailHandler := emailtemplatehttp.NewHandler(bcs.EmailTemplate, l)
	announcementHandler := announcementhttp.NewHandler(bcs.Announcement, l)
	translationHandler := translationhttp.NewHandler(bcs.Translation, l)
	siteSettingHandler := sitesettinghttp.NewHandler(bcs.SiteSetting, l)
	rateLimitHandler := ratelimithttp.NewHandler(bcs.RateLimit, l)
	ipRuleHandler := iprulehttp.NewHandler(bcs.IPRule, l)
	jobHandler := jobhttp.NewHandler(bcs.Job, l)
	dataExportHandler := dataexporthttp.NewHandler(bcs.DataExport, l)
	fileHandler := filehttp.NewHandler(bcs.File, l)
	userSettingHandler := usersettinghttp.NewHandler(bcs.UserSetting, l)
	errorCodeHandler := errorcodehttp.NewHandler(bcs.ErrorCode, l)

	// === Public routes (no auth) ===
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/sign-in", userHandler.SignIn)
		auth.POST("/sign-up", userHandler.SignUp)
	}

	// === Auth-only (sign-out) ===
	authOnly := router.Group("/api/v1/auth")
	authOnly.Use(authMW)
	{
		authOnly.POST("/sign-out", userHandler.SignOut)
	}

	// === Protected routes (auth + authz + csrf) ===

	// Users
	users := router.Group("/api/v1/users")
	users.Use(authMW, authzMW, csrfMW)
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

	// Roles
	roles := router.Group("/api/v1/roles")
	roles.Use(authMW, authzMW, csrfMW)
	{
		roles.POST("", authzHandler.CreateRole)
		roles.GET("", authzHandler.ListRoles)
		roles.GET("/:id", authzHandler.GetRole)
		roles.PATCH("/:id", authzHandler.UpdateRole)
		roles.DELETE("/:id", authzHandler.DeleteRole)
		roles.POST("/:id/permissions", authzHandler.AssignPermission)
	}

	// Permissions
	permissions := router.Group("/api/v1/permissions")
	permissions.Use(authMW, authzMW, csrfMW)
	{
		permissions.POST("", authzHandler.CreatePermission)
		permissions.GET("", authzHandler.ListPermissions)
		permissions.DELETE("/:id", authzHandler.DeletePermission)
		permissions.POST("/:id/scopes", authzHandler.AssignScope)
	}

	// Policies
	policies := router.Group("/api/v1/policies")
	policies.Use(authMW, authzMW, csrfMW)
	{
		policies.POST("", authzHandler.CreatePolicy)
		policies.GET("", authzHandler.ListPolicies)
		policies.PATCH("/:id", authzHandler.UpdatePolicy)
		policies.DELETE("/:id", authzHandler.DeletePolicy)
		policies.POST("/:id/toggle", authzHandler.TogglePolicy)
	}

	// Scopes
	scopes := router.Group("/api/v1/scopes")
	scopes.Use(authMW, authzMW, csrfMW)
	{
		scopes.POST("", authzHandler.CreateScope)
		scopes.GET("", authzHandler.ListScopes)
		scopes.DELETE("", authzHandler.DeleteScope)
	}

	// Sessions
	sessions := router.Group("/api/v1/sessions")
	sessions.Use(authMW, authzMW, csrfMW)
	{
		sessions.GET("", sessionHandler.List)
		sessions.GET("/:id", sessionHandler.Get)
	}

	// Audit Logs
	auditLogs := router.Group("/api/v1/audit-logs")
	auditLogs.Use(authMW, authzMW, csrfMW)
	{
		auditLogs.GET("", auditHandler.ListAuditLogs)
	}

	// Endpoint History
	endpointHistory := router.Group("/api/v1/endpoint-history")
	endpointHistory.Use(authMW, authzMW, csrfMW)
	{
		endpointHistory.GET("", auditHandler.ListEndpointHistory)
	}

	// Dashboard
	dashboard := router.Group("/api/v1/dashboard")
	dashboard.Use(authMW, authzMW, csrfMW)
	{
		dashboard.GET("/stats", dashboardHandler.GetStats)
	}

	// System Errors
	systemErrors := router.Group("/api/v1/system-errors")
	systemErrors.Use(authMW, authzMW, csrfMW)
	{
		systemErrors.POST("", sysErrHandler.Create)
		systemErrors.GET("", sysErrHandler.List)
		systemErrors.GET("/:id", sysErrHandler.Get)
		systemErrors.POST("/:id/resolve", sysErrHandler.Resolve)
	}

	// Metrics
	metrics := router.Group("/api/v1/metrics")
	metrics.Use(authMW, authzMW, csrfMW)
	{
		metrics.POST("", metricHandler.Create)
		metrics.GET("", metricHandler.List)
	}

	// Feature Flags
	featureFlags := router.Group("/api/v1/feature-flags")
	featureFlags.Use(authMW, authzMW, csrfMW)
	{
		featureFlags.POST("", ffHandler.Create)
		featureFlags.GET("", ffHandler.List)
		featureFlags.GET("/:id", ffHandler.Get)
		featureFlags.PATCH("/:id", ffHandler.Update)
		featureFlags.DELETE("/:id", ffHandler.Delete)
	}

	// Integrations
	integrations := router.Group("/api/v1/integrations")
	integrations.Use(authMW, authzMW, csrfMW)
	{
		integrations.POST("", integrationHandler.Create)
		integrations.GET("", integrationHandler.List)
		integrations.GET("/:id", integrationHandler.Get)
		integrations.PATCH("/:id", integrationHandler.Update)
		integrations.DELETE("/:id", integrationHandler.Delete)
	}

	// Webhooks
	webhooks := router.Group("/api/v1/webhooks")
	webhooks.Use(authMW, authzMW, csrfMW)
	{
		webhooks.POST("", webhookHandler.Create)
		webhooks.GET("", webhookHandler.List)
		webhooks.GET("/:id", webhookHandler.Get)
		webhooks.PATCH("/:id", webhookHandler.Update)
		webhooks.DELETE("/:id", webhookHandler.Delete)
	}

	// Notifications
	notifications := router.Group("/api/v1/notifications")
	notifications.Use(authMW, authzMW, csrfMW)
	{
		notifications.POST("", notifHandler.Create)
		notifications.GET("", notifHandler.List)
		notifications.GET("/:id", notifHandler.Get)
		notifications.DELETE("/:id", notifHandler.Delete)
	}

	// Email Templates
	emailTemplates := router.Group("/api/v1/email-templates")
	emailTemplates.Use(authMW, authzMW, csrfMW)
	{
		emailTemplates.POST("", emailHandler.Create)
		emailTemplates.GET("", emailHandler.List)
		emailTemplates.GET("/:id", emailHandler.Get)
		emailTemplates.PATCH("/:id", emailHandler.Update)
		emailTemplates.DELETE("/:id", emailHandler.Delete)
	}

	// Announcements
	announcements := router.Group("/api/v1/announcements")
	announcements.Use(authMW, authzMW, csrfMW)
	{
		announcements.POST("", announcementHandler.Create)
		announcements.GET("", announcementHandler.List)
		announcements.GET("/:id", announcementHandler.Get)
		announcements.PATCH("/:id", announcementHandler.Update)
		announcements.DELETE("/:id", announcementHandler.Delete)
	}

	// Translations
	translations := router.Group("/api/v1/translations")
	translations.Use(authMW, authzMW, csrfMW)
	{
		translations.POST("", translationHandler.Create)
		translations.GET("", translationHandler.List)
		translations.GET("/:id", translationHandler.Get)
		translations.PATCH("/:id", translationHandler.Update)
		translations.DELETE("/:id", translationHandler.Delete)
	}

	// Site Settings
	siteSettings := router.Group("/api/v1/site-settings")
	siteSettings.Use(authMW, authzMW, csrfMW)
	{
		siteSettings.POST("", siteSettingHandler.Create)
		siteSettings.GET("", siteSettingHandler.List)
		siteSettings.GET("/:id", siteSettingHandler.Get)
		siteSettings.PATCH("/:id", siteSettingHandler.Update)
		siteSettings.DELETE("/:id", siteSettingHandler.Delete)
	}

	// Rate Limits
	rateLimits := router.Group("/api/v1/rate-limits")
	rateLimits.Use(authMW, authzMW, csrfMW)
	{
		rateLimits.POST("", rateLimitHandler.Create)
		rateLimits.GET("", rateLimitHandler.List)
		rateLimits.GET("/:id", rateLimitHandler.Get)
		rateLimits.PATCH("/:id", rateLimitHandler.Update)
		rateLimits.DELETE("/:id", rateLimitHandler.Delete)
	}

	// IP Rules
	ipRules := router.Group("/api/v1/ip-rules")
	ipRules.Use(authMW, authzMW, csrfMW)
	{
		ipRules.POST("", ipRuleHandler.Create)
		ipRules.GET("", ipRuleHandler.List)
		ipRules.GET("/:id", ipRuleHandler.Get)
		ipRules.PATCH("/:id", ipRuleHandler.Update)
		ipRules.DELETE("/:id", ipRuleHandler.Delete)
	}

	// Jobs
	jobs := router.Group("/api/v1/jobs")
	jobs.Use(authMW, authzMW, csrfMW)
	{
		jobs.POST("", jobHandler.Create)
		jobs.GET("", jobHandler.List)
		jobs.GET("/:id", jobHandler.Get)
		jobs.PATCH("/:id", jobHandler.Update)
		jobs.DELETE("/:id", jobHandler.Delete)
	}

	// Data Exports
	dataExports := router.Group("/api/v1/data-exports")
	dataExports.Use(authMW, authzMW, csrfMW)
	{
		dataExports.POST("", dataExportHandler.Create)
		dataExports.GET("", dataExportHandler.List)
		dataExports.GET("/:id", dataExportHandler.Get)
		dataExports.PATCH("/:id", dataExportHandler.Update)
		dataExports.DELETE("/:id", dataExportHandler.Delete)
	}

	// Files
	files := router.Group("/api/v1/files")
	files.Use(authMW, authzMW, csrfMW)
	{
		files.POST("", fileHandler.Create)
		files.GET("", fileHandler.List)
		files.GET("/:id", fileHandler.Get)
	}

	// User Settings
	userSettings := router.Group("/api/v1/user-settings")
	userSettings.Use(authMW, authzMW, csrfMW)
	{
		userSettings.POST("", userSettingHandler.Upsert)
		userSettings.GET("", userSettingHandler.List)
		userSettings.DELETE("/:id", userSettingHandler.Delete)
	}

	// Error Codes
	errorCodes := router.Group("/api/v1/error-codes")
	errorCodes.Use(authMW, authzMW, csrfMW)
	{
		errorCodes.POST("", errorCodeHandler.Create)
		errorCodes.GET("", errorCodeHandler.List)
		errorCodes.GET("/:id", errorCodeHandler.Get)
		errorCodes.PATCH("/:id", errorCodeHandler.Update)
		errorCodes.DELETE("/:id", errorCodeHandler.Delete)
	}
}
