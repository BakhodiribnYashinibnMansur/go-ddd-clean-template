package app

import (
	"context"
	"fmt"

	"gct/internal/context/content/announcement"
	"gct/internal/context/iam/audit"
	"gct/internal/context/iam/authz"
	"gct/internal/context/admin/statistics"
	"gct/internal/context/admin/dataexport"
	"gct/internal/context/admin/errorcode"
	"gct/internal/context/admin/featureflag"
	"gct/internal/context/content/file"
	"gct/internal/context/admin/integration"
	"gct/internal/context/ops/iprule"

	"gct/internal/context/ops/metric"
	"gct/internal/context/content/notification"
	"gct/internal/context/ops/ratelimit"
	"gct/internal/context/iam/session"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/metrics"
	"gct/internal/context/admin/sitesetting"
	"gct/internal/context/ops/systemerror"
	"gct/internal/context/content/translation"
	"gct/internal/context/iam/user"
	"gct/internal/context/iam/user/application/command"
	"gct/internal/context/iam/usersetting"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DDDBoundedContexts holds all DDD bounded contexts.
type DDDBoundedContexts struct {
	User         *user.BoundedContext
	Authz        *authz.BoundedContext
	Session      *session.BoundedContext
	Audit        *audit.BoundedContext
	Statistics   *statistics.BoundedContext
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
func NewDDDBoundedContexts(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, bm *metrics.BusinessMetrics, jwtCfg command.JWTConfig) (*DDDBoundedContexts, error) {
	_ = bm // available for BC injection when needed
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
		Statistics: statistics.NewBoundedContext(pool, l),
	}, nil
}
