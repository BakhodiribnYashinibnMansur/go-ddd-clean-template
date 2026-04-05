// Package queue provides background job processing using Asynq.
package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/hibiken/asynq"
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

	log.Info("✅ Asynq client initialized",
		"redis_addr", redisAddr,
		"redis_db", redisDB,
	)

	return &Client{
		client: client,
		log:    log,
	}
}

// Close closes the client connection.
func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("asynq.Client.Close: %w", err)
	}
	return nil
}

// EnqueueTask enqueues a task with the given type and payload.
func (c *Client) EnqueueTask(
	ctx context.Context,
	taskType string,
	payload any,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.log.Errorc(ctx, "❌ Failed to marshal task payload",
			"task_type", taskType,
			"error", err,
		)
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, payloadBytes, opts...)
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		c.log.Errorc(ctx, "❌ Failed to enqueue task",
			"task_type", taskType,
			"error", err,
		)
		return nil, fmt.Errorf("enqueue task: %w", err)
	}

	c.log.Infoc(ctx, "📤 Task enqueued successfully",
		"task_type", taskType,
		"task_id", info.ID,
		"queue", info.Queue,
	)

	return info, nil
}

// EnqueueTaskIn enqueues a task to be processed after a delay.
func (c *Client) EnqueueTaskIn(
	ctx context.Context,
	taskType string,
	payload any,
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
	payload any,
	processAt time.Time,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.ProcessAt(processAt))
	return c.EnqueueTask(ctx, taskType, payload, opts...)
}
