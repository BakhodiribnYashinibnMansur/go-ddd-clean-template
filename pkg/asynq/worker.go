package asynq

import (
	"context"
	"fmt"

	"gct/config"
	"gct/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Worker wraps asynq.Server for task processing.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	log    logger.Log
}

// NewWorker creates a new Asynq worker.
func NewWorker(cfg config.AsynqConfig, log logger.Log) *Worker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      cfg.GetDefaultQueues(),
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.WithContext(ctx).Errorw("task processing failed",
					zap.String("task_type", task.Type()),
					zap.String("task_id", task.ResultWriter().TaskID()),
					zap.Error(err),
				)
			}),
			Logger: NewAsynqLogger(log),
		},
	)

	return &Worker{
		server: server,
		mux:    asynq.NewServeMux(),
		log:    log,
	}
}

// RegisterHandler registers a task handler.
func (w *Worker) RegisterHandler(taskType string, handler func(context.Context, *asynq.Task) error) {
	w.mux.HandleFunc(taskType, handler)
	w.log.WithContext(context.Background()).Infow("registered task handler",
		zap.String("task_type", taskType),
	)
}

// Start starts the worker server.
func (w *Worker) Start() error {
	w.log.WithContext(context.Background()).Infow("starting asynq worker")
	if err := w.server.Start(w.mux); err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	return nil
}

// Stop gracefully stops the worker server.
func (w *Worker) Stop() {
	w.log.WithContext(context.Background()).Infow("stopping asynq worker")
	w.server.Stop()
	w.server.Shutdown()
}

// AsynqLogger adapts our logger to asynq.Logger interface.
type AsynqLogger struct {
	log logger.Log
}

// NewAsynqLogger creates a new asynq logger adapter.
func NewAsynqLogger(log logger.Log) *AsynqLogger {
	return &AsynqLogger{log: log}
}

func (l *AsynqLogger) Debug(args ...interface{}) {
	l.log.WithContext(context.Background()).Debugw(fmt.Sprint(args...))
}

func (l *AsynqLogger) Info(args ...interface{}) {
	l.log.WithContext(context.Background()).Infow(fmt.Sprint(args...))
}

func (l *AsynqLogger) Warn(args ...interface{}) {
	l.log.WithContext(context.Background()).Warnw(fmt.Sprint(args...))
}

func (l *AsynqLogger) Error(args ...interface{}) {
	l.log.WithContext(context.Background()).Errorw(fmt.Sprint(args...))
}

func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.log.WithContext(context.Background()).Fatalw(fmt.Sprint(args...))
}
