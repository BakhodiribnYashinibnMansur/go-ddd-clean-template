package usersetting

import (
	"gct/internal/repo/persistent/postgres/user/setting"
	"gct/internal/shared/infrastructure/logger"
)

const (
	KeyPasscode        = "passcode"
	KeyPasscodeEnabled = "passcode_enabled"
)

type UseCase struct {
	repo   setting.RepoI
	logger logger.Log
}

func New(repo setting.RepoI, logger logger.Log) UseCaseI {
	return &UseCase{repo: repo, logger: logger}
}
