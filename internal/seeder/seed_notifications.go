package seeder

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedNotifications(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding notifications...", zap.Int("count", count))

	now := time.Now()

	types := []string{"info", "warning", "alert"}
	targetTypes := []string{"all", "admin", "user"}

	titles := []string{
		"New login detected", "Password changed successfully", "Your export is ready",
		"Account verification required", "Session expired", "New feature available",
		"Rate limit warning", "Maintenance scheduled", "Role updated",
		"File uploaded successfully", "API key expiring soon", "Security alert",
	}

	bodies := []string{
		"A new login was detected from a new device. If this wasn't you, please change your password.",
		"Your password has been updated successfully. You can now use it to log in.",
		"Your data export has been processed and is ready for download.",
		"Please verify your account to access all features.",
		"Your session has expired. Please log in again to continue.",
		"We have released a new feature. Check it out in your dashboard.",
		"You are approaching the rate limit for API requests.",
		"System maintenance is scheduled. Some services may be temporarily unavailable.",
		"Your role has been updated. You may have new permissions.",
		"Your file has been uploaded and is now available.",
		"One of your API keys is expiring soon. Please renew it.",
		"Unusual activity has been detected on your account.",
	}

	for i := 0; i < count; i++ {
		titleIdx := gofakeit.Number(0, len(titles)-1)

		_, err := s.pool.Exec(ctx,
			`INSERT INTO notifications (id, title, body, type, target_type, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.New(), titles[titleIdx], bodies[titleIdx%len(bodies)],
			types[gofakeit.Number(0, len(types)-1)], targetTypes[gofakeit.Number(0, len(targetTypes)-1)],
			true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create notification", zap.Error(err))
		}
	}

	return nil
}
