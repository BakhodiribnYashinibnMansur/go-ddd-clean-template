package admin

import (
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

func (h *Handler) NewAsynqMonitor() http.Handler {
	opts := asynqmon.Options{
		RootPath: "/admin/asynq",
		RedisConnOpt: asynq.RedisClientOpt{
			Addr:     h.cfg.Redis.Addr(),
			Password: h.cfg.Redis.Password,
			DB:       h.cfg.Redis.DB,
		},
	}
	return asynqmon.New(opts)
}
