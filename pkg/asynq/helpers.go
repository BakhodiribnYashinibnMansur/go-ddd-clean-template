package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// TaskOptions provides convenient task options.
type TaskOptions struct {
	Queue     string
	MaxRetry  int
	Timeout   time.Duration
	Deadline  time.Time
	UniqueKey string
	UniqueTTL time.Duration
	ProcessIn time.Duration
	ProcessAt time.Time
	Retention time.Duration
}

// BuildOptions converts TaskOptions to asynq.Option slice.
func (opts TaskOptions) BuildOptions() []asynq.Option {
	var options []asynq.Option

	if opts.Queue != "" {
		options = append(options, asynq.Queue(opts.Queue))
	}

	if opts.MaxRetry > 0 {
		options = append(options, asynq.MaxRetry(opts.MaxRetry))
	}

	if opts.Timeout > 0 {
		options = append(options, asynq.Timeout(opts.Timeout))
	}

	if !opts.Deadline.IsZero() {
		options = append(options, asynq.Deadline(opts.Deadline))
	}

	if opts.UniqueKey != "" {
		options = append(options, asynq.TaskID(opts.UniqueKey))
	}

	if opts.UniqueTTL > 0 {
		options = append(options, asynq.Unique(opts.UniqueTTL))
	}

	if opts.ProcessIn > 0 {
		options = append(options, asynq.ProcessIn(opts.ProcessIn))
	}

	if !opts.ProcessAt.IsZero() {
		options = append(options, asynq.ProcessAt(opts.ProcessAt))
	}

	if opts.Retention > 0 {
		options = append(options, asynq.Retention(opts.Retention))
	}

	return options
}

// Helper functions for common task enqueueing patterns

// EnqueueEmail enqueues an email task.
func (c *Client) EnqueueEmail(ctx context.Context, taskType string, payload EmailPayload, opts ...TaskOptions) (*asynq.TaskInfo, error) {
	var options []asynq.Option
	if len(opts) > 0 {
		options = opts[0].BuildOptions()
	} else {
		// Default to critical queue for emails
		options = []asynq.Option{asynq.Queue(QueueCritical)}
	}

	return c.EnqueueTask(ctx, taskType, payload, options...)
}

// EnqueueImage enqueues an image processing task.
func (c *Client) EnqueueImage(ctx context.Context, taskType string, payload ImagePayload, opts ...TaskOptions) (*asynq.TaskInfo, error) {
	var options []asynq.Option
	if len(opts) > 0 {
		options = opts[0].BuildOptions()
	} else {
		// Default to low queue for image processing
		options = []asynq.Option{asynq.Queue(QueueLow)}
	}

	return c.EnqueueTask(ctx, taskType, payload, options...)
}

// EnqueueNotification enqueues a notification task.
func (c *Client) EnqueueNotification(ctx context.Context, taskType string, payload NotificationPayload, opts ...TaskOptions) (*asynq.TaskInfo, error) {
	var options []asynq.Option
	if len(opts) > 0 {
		options = opts[0].BuildOptions()
	} else {
		// Default to default queue for notifications
		options = []asynq.Option{asynq.Queue(QueueDefault)}
	}

	return c.EnqueueTask(ctx, taskType, payload, options...)
}

// EnqueueSeed enqueues a seeding task.
func (c *Client) EnqueueSeed(ctx context.Context, payload SeedPayload, opts ...TaskOptions) (*asynq.TaskInfo, error) {
	var options []asynq.Option
	if len(opts) > 0 {
		options = opts[0].BuildOptions()
	} else {
		options = []asynq.Option{
			asynq.Queue(QueueLow),
			asynq.MaxRetry(1),
			asynq.Timeout(10 * time.Minute),
		}
	}

	return c.EnqueueTask(ctx, TypeSystemSeed, payload, options...)
}

// EnqueueAudit enqueues an audit log task.
func (c *Client) EnqueueAudit(ctx context.Context, payload AuditPayload, opts ...TaskOptions) (*asynq.TaskInfo, error) {
	var options []asynq.Option
	if len(opts) > 0 {
		options = opts[0].BuildOptions()
	} else {
		// Default to low queue for audit logging
		options = []asynq.Option{asynq.Queue(QueueLow)}
	}

	return c.EnqueueTask(ctx, TypeAuditLog, payload, options...)
}
