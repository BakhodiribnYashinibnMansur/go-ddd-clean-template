package seeder

import (
	"context"
	"fmt"
	"time"

	"gct/internal/domain"
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

	for _, name := range predefined {
		p := &domain.Permission{
			ID:        uuid.New(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Persistent.Postgres.Authz.Permission.Create(ctx, p); err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined permission", zap.Error(err), zap.String("name", name))
		}
	}

	// Add some random ones
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s.%s", gofakeit.FileExtension(), gofakeit.Verb())
		p := &domain.Permission{
			ID:        uuid.New(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Persistent.Postgres.Authz.Permission.Create(ctx, p); err != nil {
			s.logger.Warnc(ctx, "Failed to create random permission", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}

func (s *Seeder) seedRoles(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding roles...", zap.Int("count", count))

	predefined := []string{"Admin", "Manager", "User", "Auditor", "Support"}
	for _, name := range predefined {
		r := &domain.Role{
			ID:        uuid.New(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Persistent.Postgres.Authz.Role.Create(ctx, r); err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined role", zap.Error(err), zap.String("name", name))
		}
	}

	for i := 0; i < count; i++ {
		name := gofakeit.JobLevel() + " " + gofakeit.JobTitle()
		r := &domain.Role{
			ID:        uuid.New(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Persistent.Postgres.Authz.Role.Create(ctx, r); err != nil {
			s.logger.Warnc(ctx, "Failed to create random role", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}

func (s *Seeder) seedRolePermissions(ctx context.Context) error {
	s.logger.Infoc(ctx, "Seeding role-permission mappings...")

	roles, _, err := s.repo.Persistent.Postgres.Authz.Role.Gets(ctx, &domain.RolesFilter{})
	if err != nil {
		return err
	}

	perms, _, err := s.repo.Persistent.Postgres.Authz.Permission.Gets(ctx, &domain.PermissionsFilter{})
	if err != nil {
		return err
	}

	if len(roles) == 0 || len(perms) == 0 {
		return nil
	}

	for _, role := range roles {
		// Assign 1-5 random permissions to each role
		numPerms := gofakeit.Number(1, 4)
		if numPerms > len(perms) {
			numPerms = len(perms)
		}

		for i := 0; i < numPerms; i++ {
			perm := perms[gofakeit.Number(0, len(perms)-1)]
			if err := s.repo.Persistent.Postgres.Authz.Role.AddPermission(ctx, role.ID, perm.ID); err != nil {
				s.logger.Warnc(ctx, "Failed to link role and permission",
					zap.Error(err),
					zap.String("role", role.Name),
					zap.String("permission", perm.Name),
				)
			}
		}
	}

	return nil
}

func (s *Seeder) seedPolicies(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding policies...", zap.Int("count", count))

	perms, _, err := s.repo.Persistent.Postgres.Authz.Permission.Gets(ctx, &domain.PermissionsFilter{})
	if err != nil {
		return err
	}

	if len(perms) == 0 {
		return nil
	}

	for i := 0; i < count; i++ {
		perm := perms[gofakeit.Number(0, len(perms)-1)]
		p := &domain.Policy{
			ID:           uuid.New(),
			PermissionID: perm.ID,
			Effect:       domain.PolicyEffectAllow,
			Priority:     gofakeit.Number(1, 100),
			Active:       true,
			Conditions: map[string]any{
				"region": gofakeit.State(),
				"branch": gofakeit.Company(),
			},
			CreatedAt: time.Now(),
		}
		if err := s.repo.Persistent.Postgres.Authz.Policy.Create(ctx, p); err != nil {
			s.logger.Warnc(ctx, "Failed to create policy", zap.Error(err))
		}
	}

	return nil
}
