package usecase

import (
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase/user"
	"github.com/evrone/go-clean-template/pkg/logger"
)

// UseCase -.
type UseCase struct {
	User *user.User
}

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log) *UseCase {
	return &UseCase{
		User: user.New(repos, logger),
	}
}
