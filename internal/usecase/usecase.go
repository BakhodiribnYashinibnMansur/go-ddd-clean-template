package usecase

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase/audit"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/database"
	errorcode "gct/internal/usecase/errorcode"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/sitesetting"
	"gct/internal/usecase/user"
	"gct/pkg/asynq"
	"gct/pkg/logger"

	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	Repo        *repo.Repo
	User        *user.UseCase
	Minio       *minio.UseCase
	Authz       *authz.UseCase
	Audit       *audit.UseCase
	SiteSetting *sitesetting.UseCase
	ErrorCode   *errorcode.UseCase
	Database    *database.UseCase
	AsynqClient *asynq.Client
}

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log, cfg *config.Config, asynqClient *asynq.Client) *UseCase {
	return &UseCase{
		Repo:        repos,
		User:        user.New(repos, logger, cfg),
		Minio:       minio.New(repos, logger),
		Authz:       authz.New(repos, logger, cfg),
		Audit:       audit.New(repos.Persistent, logger),
		SiteSetting: sitesetting.New(repos.Persistent, logger),
		ErrorCode:   errorcode.New(repos, logger),
		Database:    database.New(repos.Persistent.Postgres, logger, cfg),
		AsynqClient: asynqClient,
	}
}

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

// LogAction records a specific business action to the audit log.
func (u *UseCase) LogAction(ctx context.Context, action domain.AuditActionType, userID *uuid.UUID, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	al := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		Success:      true,
		CreatedAt:    time.Now(),
	}

	// Async save
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = u.Audit.Log.Create(bgCtx, al)
	}()
}
