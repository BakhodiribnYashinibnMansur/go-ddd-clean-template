package tasks

import (
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeSendFCM      = "task:send_fcm"
	TypeSendFCMMulti = "task:send_fcm_multi"
	TypeSendTelegram = "task:send_telegram"
)

func DefaultRetryOpts() []asynq.Option {
	return []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.Queue("external"),
	}
}
