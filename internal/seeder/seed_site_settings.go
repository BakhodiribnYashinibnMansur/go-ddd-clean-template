package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedSiteSettings(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding site settings...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		key       string
		value     string
		valueType string
		category  string
		desc      string
		isPublic  bool
	}{
		{"site_name", "My Application", "string", "general", "The name of the site", true},
		{"site_description", "A modern web application", "string", "general", "Site description for SEO", true},
		{"maintenance_mode", "false", "bool", "general", "Enable maintenance mode", false},
		{"max_upload_size", "10485760", "int", "general", "Max file upload size in bytes", false},
		{"smtp_host", "smtp.example.com", "string", "email", "SMTP server hostname", false},
		{"smtp_port", "587", "int", "email", "SMTP server port", false},
		{"smtp_from", "noreply@example.com", "string", "email", "Default sender email", false},
		{"session_timeout", "3600", "int", "security", "Session timeout in seconds", false},
		{"rate_limit_enabled", "true", "bool", "security", "Enable global rate limiting", false},
		{"api_version", "v1", "string", "api", "Current API version", true},
	}

	for _, s2 := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO site_settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), s2.key, s2.value, s2.valueType, s2.category, s2.desc, s2.isPublic, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined site setting", zap.Error(err), zap.String("key", s2.key))
		}
	}

	categories := []string{"general", "email", "security", "api"}
	valueTypes := []string{"string", "bool", "int"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		key := fmt.Sprintf("custom_%s_%d", gofakeit.Word(), i)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO site_settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), key, gofakeit.Word(), valueTypes[gofakeit.Number(0, len(valueTypes)-1)],
			categories[gofakeit.Number(0, len(categories)-1)], gofakeit.Sentence(5), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random site setting", zap.Error(err), zap.String("key", key))
		}
	}

	return nil
}
