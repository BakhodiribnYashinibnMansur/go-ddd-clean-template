package seeder

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedIntegrations(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding integrations...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		name    string
		desc    string
		baseURL string
	}{
		{"Slack", "Slack notifications integration", "https://hooks.slack.com/services"},
		{"Telegram", "Telegram bot integration", "https://api.telegram.org"},
		{"Email SMTP", "Email notification service", "https://smtp.example.com"},
		{"Webhook Relay", "Generic webhook relay service", "https://webhook.example.com"},
		{"Monitoring API", "External monitoring integration", "https://monitor.example.com/api"},
	}

	for i, intg := range predefined {
		if i >= count {
			break
		}
		integrationID := uuid.New()
		_, err := s.pool.Exec(ctx,
			`INSERT INTO integrations (id, name, description, base_url, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			integrationID, intg.name, intg.desc, intg.baseURL, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create integration", zap.Error(err), zap.String("name", intg.name))
			continue
		}

		// Create an API key for each integration
		apiKey := generateAPIKey()
		prefix := apiKey[:8]
		_, err = s.pool.Exec(ctx,
			`INSERT INTO api_keys (id, integration_id, name, key, key_prefix, is_active, expires_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), integrationID, fmt.Sprintf("%s API Key", intg.name), apiKey, prefix, true, now.AddDate(1, 0, 0), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create API key", zap.Error(err), zap.String("integration", intg.name))
		}
	}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		integrationID := uuid.New()
		name := fmt.Sprintf("%s Integration %d", gofakeit.Company(), i)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO integrations (id, name, description, base_url, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			integrationID, name, gofakeit.Sentence(5), fmt.Sprintf("https://%s.example.com/api", gofakeit.Word()),
			gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random integration", zap.Error(err), zap.String("name", name))
			continue
		}

		apiKey := generateAPIKey()
		_, _ = s.pool.Exec(ctx,
			`INSERT INTO api_keys (id, integration_id, name, key, key_prefix, is_active, expires_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), integrationID, fmt.Sprintf("%s Key", name), apiKey, apiKey[:8], true, now.AddDate(1, 0, 0), now, now,
		)
	}

	return nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return "gct_" + hex.EncodeToString(bytes)
}
