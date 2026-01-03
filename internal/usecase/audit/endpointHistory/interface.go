package endpointHistory

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.EndpointHistory) error
	Gets(ctx context.Context, in *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error)
}
