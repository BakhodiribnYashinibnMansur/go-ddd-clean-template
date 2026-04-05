package query

import (
	"context"
	"strconv"

	"gct/internal/context/admin/supporting/sitesetting/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UserMaxSessionsKey is the site_settings key that stores the per-user max
// concurrent active session count.
const UserMaxSessionsKey = "user.max_sessions"

// DefaultUserMaxSessions is the fallback when the setting is missing, blank
// or malformed. Chosen to match the product default at time of writing.
const DefaultUserMaxSessions = 3

// GetUserMaxSessionsHandler resolves the "user.max_sessions" site setting as a
// positive integer. It degrades gracefully: any repo/parse failure is logged
// at warn-level and the default (3) is returned — sign-in must never break
// because of a misconfigured row.
type GetUserMaxSessionsHandler struct {
	readRepo domain.SiteSettingReadRepository
	logger   logger.Log
}

// NewGetUserMaxSessionsHandler creates a new GetUserMaxSessionsHandler.
func NewGetUserMaxSessionsHandler(readRepo domain.SiteSettingReadRepository, l logger.Log) *GetUserMaxSessionsHandler {
	return &GetUserMaxSessionsHandler{readRepo: readRepo, logger: l}
}

// Handle returns the configured max-sessions value, or DefaultUserMaxSessions
// if the key is missing, blank, non-numeric or <= 0. The error return is
// reserved for truly unexpected situations — callers typically ignore it.
func (h *GetUserMaxSessionsHandler) Handle(ctx context.Context) (result int, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetUserMaxSessionsHandler.Handle")
	defer func() { end(err) }()

	key := UserMaxSessionsKey
	views, _, lerr := h.readRepo.List(ctx, domain.SiteSettingFilter{Key: &key, Limit: 1})
	if lerr != nil {
		h.logger.Warnc(ctx, "max sessions lookup failed",
			logger.F{Op: "GetUserMaxSessions", Entity: "site_setting", Err: lerr}.KV()...)
		return DefaultUserMaxSessions, nil
	}
	if len(views) == 0 {
		return DefaultUserMaxSessions, nil
	}

	v := views[0].Value
	if v == "" {
		return DefaultUserMaxSessions, nil
	}
	n, perr := strconv.Atoi(v)
	if perr != nil || n <= 0 {
		h.logger.Warnc(ctx, "max sessions value invalid, using default",
			logger.F{Op: "GetUserMaxSessions", Entity: "site_setting", Err: perr}.KV()...)
		return DefaultUserMaxSessions, nil
	}
	return n, nil
}
