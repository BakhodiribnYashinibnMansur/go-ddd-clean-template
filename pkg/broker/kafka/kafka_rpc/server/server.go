// Package server implements Kafka RPC server.
package server

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	kafkarpc "gct/pkg/broker/kafka/kafka_rpc"
	"gct/pkg/logger"
)

const (
	_defaultTimeout = 2 * time.Second
)

// CallHandler -.
type CallHandler func(kafka.Message) (any, error)

// Server -.
type Server struct {
	ctx context.Context
	eg  *errgroup.Group

	reader *kafka.Reader
	writer *kafka.Writer
	router map[string]CallHandler
	stop   chan struct{}
	notify chan error

	timeout time.Duration

	logger logger.Log
}

// New -.
func New(
	brokers []string,
	topic string,
	groupId string,
	router map[string]CallHandler,
	l logger.Log,
	opts ...Option,
) (*Server, error) {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupId,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	s := &Server{
		ctx:     ctx,
		eg:      group,
		reader:  reader,
		writer:  writer,
		router:  router,
		stop:    make(chan struct{}),
		notify:  make(chan error, 1),
		timeout: _defaultTimeout,
		logger:  l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		for {
			select {
			case <-s.stop:
				return nil
			case <-s.ctx.Done():
				return s.ctx.Err()
			default:
				msg, err := s.reader.ReadMessage(s.ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					s.logger.Errorw("kafka_rpc server - Start - s.reader.ReadMessage", zap.Error(err))
					continue
				}

				go s.handleMessage(msg)
			}
		}
	})

	s.logger.Infow("kafka_rpc server - Started")
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	close(s.stop)

	var shutdownErrors []error

	if err := s.eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		shutdownErrors = append(shutdownErrors, err)
	}

	if err := s.reader.Close(); err != nil {
		shutdownErrors = append(shutdownErrors, err)
	}

	if err := s.writer.Close(); err != nil {
		shutdownErrors = append(shutdownErrors, err)
	}

	s.logger.Infow("kafka_rpc server - Shutdown")

	return errors.Join(shutdownErrors...)
}

func (s *Server) handleMessage(msg kafka.Message) {
	var handlerName string
	for _, h := range msg.Headers {
		if h.Key == "Handler" {
			handlerName = string(h.Value)
			break
		}
	}

	handler, ok := s.router[handlerName]
	if !ok {
		s.publishResponse(msg, nil, kafkarpc.ErrBadHandler)
		return
	}

	response, err := handler(msg)
	if err != nil {
		s.logger.Errorw("kafka_rpc server - handleMessage - handler", zap.Error(err))
		s.publishResponse(msg, nil, kafkarpc.ErrInternalServer)
		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		s.logger.Errorw("kafka_rpc server - handleMessage - json.Marshal", zap.Error(err))
		s.publishResponse(msg, nil, kafkarpc.ErrInternalServer)
		return
	}

	s.publishResponse(msg, body, kafkarpc.Success)
}

func (s *Server) publishResponse(req kafka.Message, body []byte, status string) {
	var replyTopic string
	var correlationID string

	for _, h := range req.Headers {
		switch h.Key {
		case "ReplyTopic":
			replyTopic = string(h.Value)
		case "CorrelationID":
			correlationID = string(h.Value)
		}
	}

	if replyTopic == "" {
		return
	}

	res := kafka.Message{
		Topic: replyTopic,
		Value: body,
		Headers: []kafka.Header{
			{Key: "Status", Value: []byte(status)},
			{Key: "CorrelationID", Value: []byte(correlationID)},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.writer.WriteMessages(ctx, res); err != nil {
		s.logger.Errorw("kafka_rpc server - publishResponse - s.writer.WriteMessages", zap.Error(err))
	}
}
