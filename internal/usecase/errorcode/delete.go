package errorcode

import (
	"context"
)

func (uc *UseCase) Delete(ctx context.Context, code string) error {
	return uc.repo.Delete(ctx, code)
}
