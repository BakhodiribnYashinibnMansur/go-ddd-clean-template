package v1

import (
	"context"
	"fmt"

	v1 "github.com/evrone/go-clean-template/docs/proto/v1"
	"github.com/evrone/go-clean-template/internal/controller/grpc/v1/response"
	"go.uber.org/zap"
)

func (r *V1) GetHistory(ctx context.Context, _ *v1.GetHistoryRequest) (*v1.GetHistoryResponse, error) {
	translationHistory, err := r.t.History(ctx)
	if err != nil {
		r.l.GetZap().Error("grpc - v1 - GetHistory",
			zap.Error(err),
		)

		return nil, fmt.Errorf("grpc - v1 - GetHistory: %w", err)
	}

	return response.NewTranslationHistory(translationHistory), nil
}
