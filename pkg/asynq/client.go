// Package queue provides background job processing using Asynq.
package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gct/pkg/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wraps asynq.Client for task enqueueing.
type Client struct {
	client *asynq.Client
	log    logger.Log
}

// NewClient creates a new Asynq client.
func NewClient(redisAddr, redisPassword string, redisDB int, log logger.Log) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &Client{
		client: client,
		log:    log,
	}
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueTask enqueues a task with the given type and payload.
func (c *Client) EnqueueTask(
	ctx context.Context,
	taskType string,
	payload interface{},
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.log.WithContext(ctx).Errorw("failed to marshal task payload",
			zap.String("task_type", taskType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, payloadBytes, opts...)
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		c.log.WithContext(ctx).Errorw("failed to enqueue task",
			zap.String("task_type", taskType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("enqueue task: %w", err)
	}

	c.log.WithContext(ctx).Infow("task enqueued successfully",
		zap.String("task_type", taskType),
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
	)

	return info, nil
}

// EnqueueTaskIn enqueues a task to be processed after a delay.
func (c *Client) EnqueueTaskIn(
	ctx context.Context,
	taskType string,
	payload interface{},
	delay time.Duration,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessIn(delay))
	return c.EnqueueTask(ctx, taskType, payload, opts...)
}

// EnqueueTaskAt enqueues a task to be processed at a specific time.
func (c *Client) EnqueueTaskAt(
	ctx context.Context,
	taskType string,
	payload interface{},
	processAt time.Time,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessAt(processAt))
	return c.EnqueueTask(ctx, taskType, payload, opts...)
}
