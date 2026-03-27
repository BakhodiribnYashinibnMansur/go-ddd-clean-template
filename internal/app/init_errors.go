package app

import (
	"context"

	"gct/internal/errorcode"
	"gct/internal/errorcode/application/command"
	"gct/internal/errorcode/application/query"
	"gct/internal/errorcode/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
)

// initErrorCodes loads dynamic error codes from the database and configures the error package.
// It also synchronizes missing error codes from the codebase to the database.
func initErrorCodes(ctx context.Context, ec *errorcode.BoundedContext, l logger.Log) {
	l.Info("Initializing and synchronizing error codes...")

	// 1. Load existing codes from DB
	result, err := ec.ListErrorCodes.Handle(ctx, query.ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{Limit: 10000},
	})
	if err != nil {
		l.Errorc(ctx, "failed to load error codes from database", "error", err)
		return
	}

	// Create a map for quick lookup
	dbCodeMap := make(map[string]bool)
	for _, c := range result.ErrorCodes {
		dbCodeMap[c.Code] = true

		// Apply DB configuration to runtime
		apperrors.ConfigureError(c.Code, apperrors.ErrorDetailConfig{
			Message: apperrors.UserMessage{
				En: c.Message,
			},
			HTTPStatus: c.HTTPStatus,
		})
	}
	l.Infoc(ctx, "loaded existing error codes from database", "count", len(result.ErrorCodes))

	// 2. Synchronize: Add missing codes from Codebase to DB
	allErrors := apperrors.GetAllErrors()
	newCount := 0

	for _, def := range allErrors {
		if !dbCodeMap[def.Code] {
			httpStatus := def.HTTPStatus
			if httpStatus == 0 {
				httpStatus = 500
			}

			cmd := command.CreateErrorCodeCommand{
				Code:       def.Code,
				Message:    def.Message,
				HTTPStatus: httpStatus,
			}

			if err := ec.CreateErrorCode.Handle(ctx, cmd); err != nil {
				l.Errorc(ctx, "failed to sync new error code to database", "code", def.Code, "error", err)
			} else {
				newCount++
			}
		}
	}

	if newCount > 0 {
		l.Infoc(ctx, "synchronized new error codes to database", "new_count", newCount)
	} else {
		l.Info("database error codes are up to date")
	}
}
