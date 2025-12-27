package v1

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/broker/nats/nats_rpc/server"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (r *V1) getHistory() server.CallHandler {
	return func(_ *nats.Msg) (interface{}, error) {
		translationHistory, err := r.t.History(context.Background())
		if err != nil {
			r.l.Errorw("nats_rpc - V1 - getHistory", zap.Error(err))

			return nil, fmt.Errorf("nats_rpc - V1 - getHistory: %w", err)
		}

		return translationHistory, nil
	}
}
