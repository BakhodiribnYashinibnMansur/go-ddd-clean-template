package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedPermissions(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding permissions...", zap.Int("count", count))

	// Some predefined permissions for realism
	predefined := []string{
		"users.view", "users.create", "users.edit", "users.delete",
		"roles.view", "roles.edit",
		"audit.view",
		"reports.generate",
	}

	now := time.Now()
	for _, name := range predefined {
		_, err := s.pool.Exec(ctx,
			"INSERT INTO permission (id, name, created_at) VALUES ($1, $2, $3)",
			uuid.New(), name, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined permission", zap.Error(err), zap.String("name", name))
		}
	}

	// Add some random ones
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s.%s", gofakeit.FileExtension(), gofakeit.Verb())
		_, err := s.pool.Exec(ctx,
			"INSERT INTO permission (id, name, created_at) VALUES ($1, $2, $3)",
			uuid.New(), name, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random permission", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}

func (s *Seeder) seedRoles(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding roles...", zap.Int("count", count))

	now := time.Now()
	predefined := []string{"Admin", "Manager", "User", "Auditor", "Support"}
	for _, name := range predefined {
		_, err := s.pool.Exec(ctx,
			"INSERT INTO role (id, name, created_at) VALUES ($1, $2, $3)",
			uuid.New(), name, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined role", zap.Error(err), zap.String("name", name))
		}
	}

	for i := 0; i < count; i++ {
		name := gofakeit.JobLevel() + " " + gofakeit.JobTitle()
		_, err := s.pool.Exec(ctx,
			"INSERT INTO role (id, name, created_at) VALUES ($1, $2, $3)",
			uuid.New(), name, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random role", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}

func (s *Seeder) seedRolePermissions(ctx context.Context) error {
	s.logger.Infoc(ctx, "Seeding role-permission mappings...")

	// Get all roles
	roleRows, err := s.pool.Query(ctx, "SELECT id FROM role")
	if err != nil {
		return err
	}
	defer roleRows.Close()

	var roleIDs []uuid.UUID
	for roleRows.Next() {
		var id uuid.UUID
		if err := roleRows.Scan(&id); err != nil {
			return err
		}
		roleIDs = append(roleIDs, id)
	}

	// Get all permissions
	permRows, err := s.pool.Query(ctx, "SELECT id FROM permission")
	if err != nil {
		return err
	}
	defer permRows.Close()

	var permIDs []uuid.UUID
	for permRows.Next() {
		var id uuid.UUID
		if err := permRows.Scan(&id); err != nil {
			return err
		}
		permIDs = append(permIDs, id)
	}

	if len(roleIDs) == 0 || len(permIDs) == 0 {
		return nil
	}

	for _, roleID := range roleIDs {
		// Assign 1-4 random permissions to each role
		numPerms := gofakeit.Number(1, 4)
		if numPerms > len(permIDs) {
			numPerms = len(permIDs)
		}

		for i := 0; i < numPerms; i++ {
			permID := permIDs[gofakeit.Number(0, len(permIDs)-1)]
			_, err := s.pool.Exec(ctx,
				"INSERT INTO role_permission (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				roleID, permID,
			)
			if err != nil {
				s.logger.Warnc(ctx, "Failed to link role and permission", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *Seeder) seedPolicies(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding policies...", zap.Int("count", count))

	// Get all permissions
	rows, err := s.pool.Query(ctx, "SELECT id FROM permission")
	if err != nil {
		return err
	}
	defer rows.Close()

	var permIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		permIDs = append(permIDs, id)
	}

	if len(permIDs) == 0 {
		return nil
	}

	now := time.Now()
	for i := 0; i < count; i++ {
		permID := permIDs[gofakeit.Number(0, len(permIDs)-1)]
		conditions := fmt.Sprintf(`{"region":"%s","branch":"%s"}`, gofakeit.State(), gofakeit.Company())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO policy (id, permission_id, effect, priority, active, conditions, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), permID, "ALLOW", gofakeit.Number(1, 100), true, conditions, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create policy", zap.Error(err))
		}
	}

	return nil
}
