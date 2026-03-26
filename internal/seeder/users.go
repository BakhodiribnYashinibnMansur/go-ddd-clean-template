package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *Seeder) seedUsers(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding users...", zap.Int("count", count))

	// Get all role IDs to assign them to users
	rows, err := s.pool.Query(ctx, "SELECT id FROM role")
	if err != nil {
		return fmt.Errorf("failed to get roles for user seeding: %w", err)
	}
	defer rows.Close()

	var roleIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan role id: %w", err)
		}
		roleIDs = append(roleIDs, id)
	}

	// Standard password for all fake users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	for i := 0; i < count; i++ {
		id := uuid.New()
		username := gofakeit.Username()
		email := gofakeit.Email()
		phone := gofakeit.Phone()
		active := gofakeit.Bool()
		now := time.Now()

		var roleID *uuid.UUID
		if len(roleIDs) > 0 {
			r := roleIDs[gofakeit.Number(0, len(roleIDs)-1)]
			roleID = &r
		}

		attributes := fmt.Sprintf(`{"region":"%s","branch":"%s","dept":"%s"}`,
			gofakeit.State(), gofakeit.Company(), gofakeit.JobTitle())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO users (id, username, email, phone, password, role_id, active, attributes, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			id, username, email, phone, string(hashedPassword), roleID, active, attributes, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create fake user", zap.Error(err), zap.String("username", username))
			continue
		}
	}

	return nil
}

func (s *Seeder) seedUserRoles(ctx context.Context) error {
	s.logger.Infoc(ctx, "Seeding user roles association...")
	return nil
}
