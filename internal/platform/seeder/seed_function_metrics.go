package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFunctionMetrics(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding function metrics...", zap.Int("count", count))

	functionNames := []string{
		"UserService.Create", "UserService.GetByID", "UserService.Update",
		"AuthService.Login", "AuthService.Logout", "AuthService.RefreshToken",
		"RoleService.AssignRole", "PolicyService.Evaluate",
		"FileService.Upload", "FileService.Download",
		"NotificationService.Send", "AuditService.Log",
		"ExportService.GenerateCSV", "DashboardService.GetStats",
		"IntegrationService.Sync", "TranslationService.GetAll",
	}

	for i := 0; i < count; i++ {
		name := functionNames[gofakeit.Number(0, len(functionNames)-1)]
		latency := gofakeit.Number(5, 500)
		isPanic := gofakeit.Float64Range(0, 1) < 0.05 // 5% panic rate
		var panicError *string
		if isPanic {
			err := fmt.Sprintf("panic: %s at %s:%d", gofakeit.ErrorRuntime().Error(), gofakeit.Word()+".go", gofakeit.Number(10, 500))
			panicError = &err
		}
		createdAt := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO function_metrics (id, name, latency_ms, is_panic, panic_error, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), name, latency, isPanic, panicError, createdAt,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create function metric", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}
