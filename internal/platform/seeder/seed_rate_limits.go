package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedRateLimits(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding rate limits...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		name          string
		pathPattern   string
		method        string
		limitCount    int
		windowSeconds int
	}{
		{"Login Rate Limit", "/auth/login", "POST", 5, 60},
		{"Register Rate Limit", "/auth/register", "POST", 3, 300},
		{"API General Limit", "/api/v1/*", "ALL", 100, 60},
		{"File Upload Limit", "/api/v1/files/upload", "POST", 10, 300},
		{"Password Reset Limit", "/auth/reset-password", "POST", 3, 600},
		{"Export Rate Limit", "/api/v1/export/*", "POST", 5, 300},
	}

	for _, rl := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO rate_limits (id, name, path_pattern, method, limit_count, window_seconds, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), rl.name, rl.pathPattern, rl.method, rl.limitCount, rl.windowSeconds, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined rate limit", zap.Error(err), zap.String("name", rl.name))
		}
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "ALL"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		name := fmt.Sprintf("%s Limit %d", gofakeit.Word(), i)
		path := fmt.Sprintf("/api/v1/%s/*", gofakeit.Word())
		_, err := s.pool.Exec(ctx,
			`INSERT INTO rate_limits (id, name, path_pattern, method, limit_count, window_seconds, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), name, path, methods[gofakeit.Number(0, len(methods)-1)],
			gofakeit.Number(10, 200), gofakeit.Number(30, 600), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random rate limit", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}
