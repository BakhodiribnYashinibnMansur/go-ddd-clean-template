package client

import (
	"context"
)

func (uc *UseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.User.Client.Delete(ctx, id)
}
