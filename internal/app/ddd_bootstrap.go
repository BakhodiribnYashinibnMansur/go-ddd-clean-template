// Package app — DDD bootstrap wiring.
// Creates all DDD Bounded Contexts and returns them in a container.
// This runs alongside the existing app.go, it does NOT replace it.
package app

import (
	"crypto/rsa"
	"time"

	"gct/internal/announcement"
	"gct/internal/audit"
	"gct/internal/authz"
	"gct/internal/dashboard"
	"gct/internal/dataexport"
	"gct/internal/emailtemplate"
	"gct/internal/errorcode"
	"gct/internal/featureflag"
	"gct/internal/file"
	"gct/internal/integration"
	"gct/internal/iprule"
	"gct/internal/job"
	"gct/internal/metric"
	"gct/internal/notification"
	"gct/internal/ratelimit"
	"gct/internal/session"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting"
	"gct/internal/systemerror"
	"gct/internal/translation"
	"gct/internal/user"
	"gct/internal/user/application/command"
	"gct/internal/usersetting"
	"gct/internal/webhook"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DDDBoundedContexts holds all DDD bounded contexts.
type DDDBoundedContexts struct {
	User          *user.BoundedContext
	Authz         *authz.BoundedContext
	Session       *session.BoundedContext
	Audit         *audit.BoundedContext
	Dashboard     *dashboard.BoundedContext
	SystemError   *systemerror.BoundedContext
	Metric        *metric.BoundedContext
	FeatureFlag   *featureflag.BoundedContext
	Integration   *integration.BoundedContext
	Webhook       *webhook.BoundedContext
	Notification  *notification.BoundedContext
	EmailTemplate *emailtemplate.BoundedContext
	Announcement  *announcement.BoundedContext
	Translation   *translation.BoundedContext
	SiteSetting   *sitesetting.BoundedContext
	RateLimit     *ratelimit.BoundedContext
	IPRule        *iprule.BoundedContext
	Job           *job.BoundedContext
	DataExport    *dataexport.BoundedContext
	File          *file.BoundedContext
	UserSetting   *usersetting.BoundedContext
	ErrorCode     *errorcode.BoundedContext
}

// NewDDDBoundedContexts creates all DDD bounded contexts with their correct dependencies.
func NewDDDBoundedContexts(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, jwtPrivateKey *rsa.PrivateKey, jwtIssuer string, jwtAccessTTL, jwtRefreshTTL time.Duration) *DDDBoundedContexts {
	userJWTCfg := command.JWTConfig{
		PrivateKey: jwtPrivateKey,
		Issuer:     jwtIssuer,
		AccessTTL:  jwtAccessTTL,
		RefreshTTL: jwtRefreshTTL,
	}

	return &DDDBoundedContexts{
		// Full BCs: (pool, eventBus, logger)
		User:          user.NewBoundedContext(pool, eventBus, l, userJWTCfg),
		Authz:         authz.NewBoundedContext(pool, eventBus, l),
		Audit:         audit.NewBoundedContext(pool, eventBus, l),
		SystemError:   systemerror.NewBoundedContext(pool, eventBus, l),
		Metric:        metric.NewBoundedContext(pool, eventBus, l),
		FeatureFlag:   featureflag.NewBoundedContext(pool, eventBus, l),
		Integration:   integration.NewBoundedContext(pool, eventBus, l),
		Webhook:       webhook.NewBoundedContext(pool, eventBus, l),
		Notification:  notification.NewBoundedContext(pool, eventBus, l),
		EmailTemplate: emailtemplate.NewBoundedContext(pool, eventBus, l),
		Announcement:  announcement.NewBoundedContext(pool, eventBus, l),
		Translation:   translation.NewBoundedContext(pool, eventBus, l),
		SiteSetting:   sitesetting.NewBoundedContext(pool, eventBus, l),
		RateLimit:     ratelimit.NewBoundedContext(pool, eventBus, l),
		IPRule:        iprule.NewBoundedContext(pool, eventBus, l),
		Job:           job.NewBoundedContext(pool, eventBus, l),
		DataExport:    dataexport.NewBoundedContext(pool, eventBus, l),
		File:          file.NewBoundedContext(pool, eventBus, l),
		UserSetting:   usersetting.NewBoundedContext(pool, eventBus, l),
		ErrorCode:     errorcode.NewBoundedContext(pool, eventBus, l),

		// Read-only BCs: (pool, logger) — no eventBus
		Session:   session.NewBoundedContext(pool, l),
		Dashboard: dashboard.NewBoundedContext(pool, l),
	}
}
