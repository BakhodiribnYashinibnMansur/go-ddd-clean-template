package seeder

import (
	"context"
	"fmt"
	"gct/internal/domain"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedUsers(ctx context.Context, count int) error {
	s.logger.WithContext(ctx).Infow("Seeding users...", zap.Int("count", count))

	// Get all role IDs to assign them to users
	roles, _, err := s.repo.Persistent.Postgres.Authz.Role.Gets(ctx, &domain.RolesFilter{})
	if err != nil {
		return fmt.Errorf("failed to get roles for user seeding: %w", err)
	}

	for i := 0; i < count; i++ {
		u := domain.NewUser()
		u.ID = uuid.New()

		username := gofakeit.Username()
		email := gofakeit.Email()
		phone := gofakeit.Phone()

		u.Username = &username
		u.Email = &email
		u.Phone = &phone

		if len(roles) > 0 {
			role := roles[gofakeit.Number(0, len(roles)-1)]
			u.RoleID = &role.ID
		}

		// Standard password for all fake users for easier testing
		err := u.SetPassword("Password123!")
		if err != nil {
			return fmt.Errorf("failed to set password for fake user: %w", err)
		}

		u.Attributes = map[string]any{
			"region": gofakeit.State(),
			"branch": gofakeit.Company(),
			"dept":   gofakeit.JobTitle(),
		}

		u.Active = gofakeit.Bool()
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()

		if err := s.repo.Persistent.Postgres.User.Client.Create(ctx, u); err != nil {
			s.logger.WithContext(ctx).Warnw("Failed to create fake user", zap.Error(err), zap.String("username", username))
			continue
		}
	}

	return nil
}

func (s *Seeder) seedUserRoles(ctx context.Context) error {
	s.logger.WithContext(ctx).Infow("Seeding user roles association...")
	return nil
}
