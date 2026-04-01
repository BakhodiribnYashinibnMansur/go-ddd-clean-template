package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFeatureFlags(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding feature flags...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		key   string
		name  string
		fType string
		value string
		desc  string
	}{
		{"dark_mode", "Dark Mode", "bool", "false", "Enable dark mode for the UI"},
		{"new_dashboard", "New Dashboard", "bool", "true", "Show the redesigned dashboard"},
		{"beta_api_v2", "Beta API v2", "bool", "false", "Enable API v2 beta endpoints"},
		{"max_upload_mb", "Max Upload Size (MB)", "int", "50", "Maximum file upload size in megabytes"},
		{"welcome_message", "Welcome Message", "string", "Welcome to our platform!", "Landing page welcome message"},
		{"maintenance_banner", "Maintenance Banner", "string", "", "Banner text during maintenance"},
		{"signup_enabled", "User Signup", "bool", "true", "Allow new user registrations"},
		{"export_formats", "Export Formats", "json", `["csv","xlsx","pdf"]`, "Available data export formats"},
	}

	for _, ff := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO feature_flags (id, key, name, type, value, description, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), ff.key, ff.name, ff.fType, ff.value, ff.desc, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined feature flag", zap.Error(err), zap.String("key", ff.key))
		}
	}

	flagTypes := []string{"bool", "string", "int"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		key := fmt.Sprintf("%s_%s_%d", gofakeit.Word(), gofakeit.Word(), i)
		fType := flagTypes[gofakeit.Number(0, len(flagTypes)-1)]
		var value string
		switch fType {
		case "bool":
			value = fmt.Sprintf("%t", gofakeit.Bool())
		case "int":
			value = fmt.Sprintf("%d", gofakeit.Number(1, 1000))
		case "string":
			value = gofakeit.Word()
		}
		_, err := s.pool.Exec(ctx,
			`INSERT INTO feature_flags (id, key, name, type, value, description, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), key, gofakeit.Sentence(3), fType, value, gofakeit.Sentence(5), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random feature flag", zap.Error(err), zap.String("key", key))
		}
	}

	return nil
}
