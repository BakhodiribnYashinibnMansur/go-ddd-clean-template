package seeder

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedAnnouncements(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding announcements...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		title   string
		content string
		aType   string
	}{
		{"System Update v2.5", "We have released a major system update with performance improvements and new features.", "info"},
		{"Scheduled Maintenance", "The system will undergo maintenance on Saturday from 02:00 to 06:00 UTC.", "warning"},
		{"New Feature: Data Export", "You can now export your data in CSV, XLSX, and PDF formats from the dashboard.", "info"},
		{"Security Advisory", "Please update your passwords. We have enhanced our security policies.", "critical"},
		{"Welcome to the Platform", "Thank you for joining! Explore our features and let us know your feedback.", "info"},
	}

	types := []string{"info", "warning", "critical"}

	for i, ann := range predefined {
		if i >= count {
			break
		}
		startsAt := now.AddDate(0, 0, -gofakeit.Number(0, 30))
		endsAt := now.AddDate(0, 0, gofakeit.Number(1, 60))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO announcements (id, title, content, type, is_active, starts_at, ends_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), ann.title, ann.content, ann.aType, true, startsAt, endsAt, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined announcement", zap.Error(err), zap.String("title", ann.title))
		}
	}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		startsAt := gofakeit.DateRange(now.AddDate(0, -1, 0), now)
		endsAt := gofakeit.DateRange(now, now.AddDate(0, 2, 0))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO announcements (id, title, content, type, is_active, starts_at, ends_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), gofakeit.Sentence(4), gofakeit.Paragraph(1, 3, 10, " "),
			types[gofakeit.Number(0, len(types)-1)], gofakeit.Bool(), startsAt, endsAt, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random announcement", zap.Error(err))
		}
	}

	return nil
}
