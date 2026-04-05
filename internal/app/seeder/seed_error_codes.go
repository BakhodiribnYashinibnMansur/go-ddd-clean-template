package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type predefinedErrorCode struct {
	code       string
	message    string
	httpStatus int
	category   string
	severity   string
	retryable  bool
	retryAfter int
	suggestion string
}

//nolint:gochecknoglobals // static seed table
var predefinedErrorCodes = []predefinedErrorCode{
	{"AUTH_001", "Invalid credentials", 401, "AUTH", "MEDIUM", false, 0, "Check username and password"},
	{"AUTH_002", "Token expired", 401, "AUTH", "LOW", true, 0, "Refresh your access token"},
	{"AUTH_003", "Insufficient permissions", 403, "AUTH", "MEDIUM", false, 0, "Contact administrator for access"},
	{"DATA_001", "Resource not found", 404, "DATA", "LOW", false, 0, "Verify the resource ID"},
	{"DATA_002", "Duplicate entry", 409, "DATA", "LOW", false, 0, "Use a unique value"},
	{"DATA_003", "Invalid input format", 400, "VALIDATION", "LOW", false, 0, "Check the request body format"},
	{"SYS_001", "Internal server error", 500, "SYSTEM", "HIGH", true, 30, "Try again later"},
	{"SYS_002", "Service unavailable", 503, "SYSTEM", "CRITICAL", true, 60, "Service is under maintenance"},
	{"SYS_003", "Database connection failed", 500, "SYSTEM", "CRITICAL", true, 10, "Try again shortly"},
	{"BIZ_001", "Rate limit exceeded", 429, "BUSINESS", "LOW", true, 60, "Wait before retrying"},
	{"BIZ_002", "Account suspended", 403, "BUSINESS", "HIGH", false, 0, "Contact support"},
	{"VAL_001", "Required field missing", 400, "VALIDATION", "LOW", false, 0, "Provide all required fields"},
}

func (s *Seeder) seedErrorCodes(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding error codes...", zap.Int("count", count))

	now := time.Now()

	for _, ec := range predefinedErrorCodes {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO error_code (id, code, message, http_status, category, severity, retryable, retry_after, suggestion, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			uuid.New(), ec.code, ec.message, ec.httpStatus, ec.category, ec.severity, ec.retryable, ec.retryAfter, ec.suggestion, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined error code", zap.Error(err), zap.String("code", ec.code))
		}
	}

	categories := []string{"DATA", "AUTH", "SYSTEM", "VALIDATION", "BUSINESS"}
	severities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	statuses := []int{400, 401, 403, 404, 409, 422, 429, 500, 502, 503}

	for i := 0; i < count-len(predefinedErrorCodes); i++ {
		if i+len(predefinedErrorCodes) >= count {
			break
		}
		cat := categories[gofakeit.Number(0, len(categories)-1)]
		code := fmt.Sprintf("%s_%03d", cat, i+100)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO error_code (id, code, message, http_status, category, severity, retryable, retry_after, suggestion, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			uuid.New(), code, gofakeit.Sentence(4), statuses[gofakeit.Number(0, len(statuses)-1)],
			cat, severities[gofakeit.Number(0, len(severities)-1)], gofakeit.Bool(), gofakeit.Number(0, 120),
			gofakeit.Sentence(6), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random error code", zap.Error(err), zap.String("code", code))
		}
	}

	return nil
}
