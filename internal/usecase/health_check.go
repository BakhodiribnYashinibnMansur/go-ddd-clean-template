package usecase

import (
	"context"
	"fmt"
)

// HealthCheck checks the health of the application dependencies.
func (u *UseCase) HealthCheck(ctx context.Context) error {
	// Check Postgres
	if err := u.Repo.Persistent.Postgres.Ping(ctx); err != nil {
		return fmt.Errorf("postgres check failed: %w", err)
	}

	// Check Redis
	if err := u.Repo.Persistent.Redis.Ping(ctx); err != nil {
		return fmt.Errorf("redis check failed: %w", err)
	}

	return nil
}
