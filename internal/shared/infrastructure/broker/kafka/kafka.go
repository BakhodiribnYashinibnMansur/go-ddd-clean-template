// Package kafka implements Kafka producer and consumer.
package kafka

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/shared/infrastructure/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const (
	defaultDialTimeout    = 10 * time.Second
	defaultReadTimeout    = 10 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultCommitInterval = 1 * time.Second
)

// Producer wraps kafka.Writer for producing messages.
type Producer struct {
	Writer *kafka.Writer
	logger logger.Log
}

// Consumer wraps kafka.Reader for consuming messages.
type Consumer struct {
	Reader *kafka.Reader
	logger logger.Log
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg config.Kafka, l logger.Log, opts ...ProducerOption) (*Producer, error) {
	writerConfig := kafka.WriterConfig{
		Brokers:      cfg.Brokers,
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: defaultWriteTimeout,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&writerConfig)
	}

	writer := kafka.NewWriter(writerConfig)

	p := &Producer{
		Writer: writer,
		logger: l,
	}

	l.Infow("Kafka producer created successfully", zap.Strings("brokers", cfg.Brokers), zap.String("topic", cfg.Topic))

	return p, nil
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(ctx context.Context, cfg config.Kafka, l logger.Log, opts ...ConsumerOption) (*Consumer, error) {
	readerConfig := kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupId,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: defaultCommitInterval,
		StartOffset:    kafka.LastOffset,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&readerConfig)
	}

	reader := kafka.NewReader(readerConfig)

	c := &Consumer{
		Reader: reader,
		logger: l,
	}

	l.Infow("Kafka consumer created successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
		zap.String("groupId", cfg.GroupId))

	return c, nil
}

// Publish sends a message to Kafka topic.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}

	if err := p.Writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Errorw("failed to publish message to Kafka", zap.Error(err))
		return fmt.Errorf("publish kafka message: %w", err)
	}

	return nil
}

// Close gracefully closes the Kafka producer.
func (p *Producer) Close() error {
	if p != nil && p.Writer != nil {
		return p.Writer.Close()
	}
	return nil
}

// Consume reads messages from Kafka topic.
func (c *Consumer) Consume(ctx context.Context) (kafka.Message, error) {
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		c.logger.Errorw("failed to consume message from Kafka", zap.Error(err))
		return kafka.Message{}, fmt.Errorf("consume kafka message: %w", err)
	}

	return msg, nil
}

// Close gracefully closes the Kafka consumer.
func (c *Consumer) Close() error {
	if c != nil && c.Reader != nil {
		return c.Reader.Close()
	}
	return nil
}
