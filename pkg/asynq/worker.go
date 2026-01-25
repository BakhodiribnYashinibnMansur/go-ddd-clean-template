package asynq

import (
	"context"
	"fmt"

	"gct/config"
	"gct/pkg/logger"

	"github.com/hibiken/asynq"
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
				log.Errorc(ctx, "❌ Asynq task processing failed",
					"task_type", task.Type(),
					"task_id", task.ResultWriter().TaskID(),
					"error", err,
					"payload_size", len(task.Payload()),
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
}

// Start starts the worker server.
func (w *Worker) Start() error {
	if err := w.server.Start(w.mux); err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	return nil
}

// Stop gracefully stops the worker server.
func (w *Worker) Stop() {
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
	l.log.Debug(fmt.Sprint(args...))
}

func (l *AsynqLogger) Info(args ...interface{}) {
	// Show info logs with emoji for better visibility
	msg := fmt.Sprint(args...)
	if len(msg) > 0 {
		l.log.Info("ℹ️  Asynq: " + msg)
	}
}

func (l *AsynqLogger) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.log.Warn("⚠️  Asynq warning: " + msg)
}

func (l *AsynqLogger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.log.Error("❌ Asynq error: " + msg)
}

func (l *AsynqLogger) Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.log.Fatal("💀 Asynq fatal: " + msg)
}
