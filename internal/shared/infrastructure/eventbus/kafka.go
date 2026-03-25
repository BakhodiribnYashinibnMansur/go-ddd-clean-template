package eventbus

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/logger"
)

// Compile-time check that KafkaEventBus implements application.EventBus.
var _ application.EventBus = (*KafkaEventBus)(nil)

// KafkaEventBus implements application.EventBus using Kafka + Protobuf.
// TODO: Full implementation in Plan 4 (Kafka setup).
type KafkaEventBus struct {
	logger logger.Log
	// kafkaWriter *kafka.Writer  // to be added
}

func NewKafkaEventBus(l logger.Log) *KafkaEventBus {
	return &KafkaEventBus{logger: l}
}

func (b *KafkaEventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	// TODO: Serialize to Protobuf and publish to Kafka topic
	for _, event := range events {
		b.logger.Infoc(ctx, "kafka event published (stub)", "event", event.EventName(), "aggregate_id", event.AggregateID())
	}
	return nil
}

func (b *KafkaEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	// TODO: Subscribe to Kafka consumer group
	b.logger.Info("kafka event subscribed (stub)", "event", eventName)
	return nil
}
