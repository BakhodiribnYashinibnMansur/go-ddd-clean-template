package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// ProducerOption defines a function type for configuring Kafka producer.
type ProducerOption func(*kafka.WriterConfig)

// ConsumerOption defines a function type for configuring Kafka consumer.
type ConsumerOption func(*kafka.ReaderConfig)

// WithProducerAsync enables async writes.
func WithProducerAsync(async bool) ProducerOption {
	return func(cfg *kafka.WriterConfig) {
		cfg.Async = async
	}
}

// WithProducerBatchSize sets the batch size for producer.
func WithProducerBatchSize(size int) ProducerOption {
	return func(cfg *kafka.WriterConfig) {
		cfg.BatchSize = size
	}
}

// WithProducerWriteTimeout sets the write timeout.
func WithProducerWriteTimeout(d time.Duration) ProducerOption {
	return func(cfg *kafka.WriterConfig) {
		cfg.WriteTimeout = d
	}
}

// WithConsumerMinBytes sets the minimum bytes to fetch.
func WithConsumerMinBytes(n int) ConsumerOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.MinBytes = n
	}
}

// WithConsumerMaxBytes sets the maximum bytes to fetch.
func WithConsumerMaxBytes(n int) ConsumerOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.MaxBytes = n
	}
}

// WithConsumerCommitInterval sets the commit interval.
func WithConsumerCommitInterval(d time.Duration) ConsumerOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.CommitInterval = d
	}
}

// WithConsumerStartOffset sets the starting offset.
func WithConsumerStartOffset(offset int64) ConsumerOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.StartOffset = offset
	}
}
