package app

import (
	"context"
	"fmt"

	"gct/internal/announcement"
	"gct/internal/audit"
	"gct/internal/authz"
	"gct/internal/dashboard"
	"gct/internal/dataexport"
	"gct/internal/errorcode"
	"gct/internal/featureflag"
	"gct/internal/file"
	"gct/internal/integration"
	"gct/internal/iprule"

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

	"github.com/jackc/pgx/v5/pgxpool"
)

// DDDBoundedContexts holds all DDD bounded contexts.
type DDDBoundedContexts struct {
	User         *user.BoundedContext
	Authz        *authz.BoundedContext
	Session      *session.BoundedContext
	Audit        *audit.BoundedContext
	Dashboard    *dashboard.BoundedContext
	SystemError  *systemerror.BoundedContext
	Metric       *metric.BoundedContext
	FeatureFlag  *featureflag.BoundedContext
	Integration  *integration.BoundedContext

	Notification *notification.BoundedContext
	Announcement *announcement.BoundedContext
	Translation  *translation.BoundedContext
	SiteSetting  *sitesetting.BoundedContext
	RateLimit    *ratelimit.BoundedContext
	IPRule       *iprule.BoundedContext

	DataExport   *dataexport.BoundedContext
	File         *file.BoundedContext
	UserSetting  *usersetting.BoundedContext
	ErrorCode    *errorcode.BoundedContext
}

// NewDDDBoundedContexts creates all bounded contexts with their dependencies.
func NewDDDBoundedContexts(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, jwtCfg command.JWTConfig) (*DDDBoundedContexts, error) {
	ffBC, err := featureflag.NewBoundedContext(ctx, pool, eventBus, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag BC: %w", err)
	}

	return &DDDBoundedContexts{
		User:         user.NewBoundedContext(pool, eventBus, l, jwtCfg),
		Authz:        authz.NewBoundedContext(pool, eventBus, l),
		Audit:        audit.NewBoundedContext(pool, eventBus, l),
		SystemError:  systemerror.NewBoundedContext(pool, eventBus, l),
		Metric:       metric.NewBoundedContext(pool, eventBus, l),
		FeatureFlag:  ffBC,
		Integration:  integration.NewBoundedContext(pool, eventBus, l),

		Notification: notification.NewBoundedContext(pool, eventBus, l),
		Announcement: announcement.NewBoundedContext(pool, eventBus, l),
		Translation:  translation.NewBoundedContext(pool, eventBus, l),
		SiteSetting:  sitesetting.NewBoundedContext(pool, eventBus, l),
		RateLimit:    ratelimit.NewBoundedContext(pool, eventBus, l),
		IPRule:       iprule.NewBoundedContext(pool, eventBus, l),

		DataExport:   dataexport.NewBoundedContext(pool, eventBus, l),
		File:         file.NewBoundedContext(pool, eventBus, l),
		UserSetting:  usersetting.NewBoundedContext(pool, eventBus, l),
		ErrorCode:    errorcode.NewBoundedContext(pool, eventBus, l),

		// Read-only BCs — no eventBus
		Session:   session.NewBoundedContext(pool, l),
		Dashboard: dashboard.NewBoundedContext(pool, l),
	}, nil
}
