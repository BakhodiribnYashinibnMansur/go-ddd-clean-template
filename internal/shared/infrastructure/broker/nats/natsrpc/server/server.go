// Package server implements NATS RPC server.
package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	natsrpc "gct/internal/shared/infrastructure/broker/nats/natsrpc"
	"gct/internal/shared/infrastructure/logger"
	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// CallHandler -.
type CallHandler func(*nats.Msg) (any, error)

// Server -.
type Server struct {
	eg *errgroup.Group

	subject      string
	connection   *nats.Conn
	subscription *nats.Subscription
	router       map[string]CallHandler
	stop         chan struct{}
	notify       chan error

	timeout time.Duration

	logger logger.Log
}

// New -.
func New(
	url,
	serverSubject string,
	router map[string]CallHandler,
	l logger.Log,
	opts ...Option,
) (*Server, error) {
	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(1) // Run only one goroutine

	connection, err := nats.Connect(
		url,
		nats.ReconnectWait(_defaultWaitTime),
		nats.MaxReconnects(_defaultAttempts),
		nats.Timeout(_defaultWaitTime),
	)
	if err != nil {
		return nil, fmt.Errorf("nats_rpc server - NewServer - nats.Connect: %w", err)
	}

	s := &Server{
		eg:         group,
		subject:    serverSubject,
		connection: connection,
		router:     router,
		stop:       make(chan struct{}),
		notify:     make(chan error, 1),
		timeout:    _defaultTimeout,
		logger:     l,
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
		err := s.subscribe()
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		// Wait for stop signal
		<-s.stop

		return nil
	})

	s.logger.Infow("nats_rpc server - Server - Started")
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	var shutdownErrors []error

	close(s.stop)

	// Wait for all goroutines to finish and get any error
	err := s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Errorw("nats_rpc server - Server - Shutdown - s.eg.Wait", zap.Error(err))

		shutdownErrors = append(shutdownErrors, err)
	}

	// Unsubscribe
	if s.subscription != nil {
		err := s.subscription.Unsubscribe()
		if err != nil {
			s.logger.Errorw("nats_rpc server - Server - Shutdown - s.conn.Subscription.Unsubscribe", zap.Error(err))

			shutdownErrors = append(shutdownErrors, err)
		}
	}

	// Close connection
	s.connection.Close()

	s.logger.Infow("nats_rpc server - Server - Shutdown")

	return errors.Join(shutdownErrors...)
}

func (s *Server) subscribe() error {
	subscription, err := s.connection.Subscribe(s.subject, s.handleMessage)
	if err != nil {
		return fmt.Errorf("nats_rpc server - subscribe - s.conn.AttemptConnect: %w", err)
	}

	s.subscription = subscription

	return nil
}

func (s *Server) handleMessage(msg *nats.Msg) {
	handler := msg.Header.Get("Handler")

	callHandler, ok := s.router[handler]
	if !ok {
		s.publish(msg, nil, natsrpc.ErrBadHandler.Error())

		return
	}

	response, err := callHandler(msg)
	if err != nil {
		s.publish(msg, nil, natsrpc.ErrInternalServer.Error())

		s.logger.Errorw("nats_rpc server - Server - handleMessage - callHandler", zap.Error(err))

		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		s.logger.Errorw("nats_rpc server - Server - handleMessage - json.Marshal", zap.Error(err))

		s.publish(msg, nil, natsrpc.ErrInternalServer.Error())

		return
	}

	s.publish(msg, body, natsrpc.Success)
}

func (s *Server) publish(msg *nats.Msg, body []byte, status string) {
	respondMsg := nats.NewMsg(msg.Reply)
	respondMsg.Header.Set("Status", status)
	respondMsg.Data = body

	err := s.connection.PublishMsg(respondMsg)
	if err != nil {
		s.logger.Errorw("nats_rpc server - Server - publish - msg.Respond", zap.Error(err))
	}
}
