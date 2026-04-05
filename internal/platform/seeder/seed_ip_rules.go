package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedIPRules(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding IP rules...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		ip     string
		action string
		reason string
	}{
		{"127.0.0.1", "allow", "Localhost"},
		{"10.0.0.0/8", "allow", "Internal network"},
		{"192.168.0.0/16", "allow", "Private network"},
		{"203.0.113.50", "block", "Known malicious IP"},
		{"198.51.100.0/24", "block", "Suspicious subnet"},
	}

	for _, rule := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO ip_rules (id, ip_address, type, reason, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), rule.ip, rule.action, rule.reason, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined IP rule", zap.Error(err), zap.String("ip", rule.ip))
		}
	}

	actions := []string{"allow", "block"}
	reasons := []string{"Automated scan detected", "Brute force attempt", "VPN access", "Office network", "Partner API"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		ip := fmt.Sprintf("%d.%d.%d.%d", gofakeit.Number(1, 223), gofakeit.Number(0, 255), gofakeit.Number(0, 255), gofakeit.Number(1, 254))
		_, err := s.pool.Exec(ctx,
			`INSERT INTO ip_rules (id, ip_address, type, reason, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), ip, actions[gofakeit.Number(0, 1)], reasons[gofakeit.Number(0, len(reasons)-1)], gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random IP rule", zap.Error(err), zap.String("ip", ip))
		}
	}

	return nil
}
