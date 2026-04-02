package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type TelegramSender interface {
	Send(msgType, text string) error
}

type TelegramPayload struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

type TelegramHandler struct {
	sender TelegramSender
}

func NewTelegramHandler(sender TelegramSender) *TelegramHandler {
	return &TelegramHandler{sender: sender}
}

func (h *TelegramHandler) HandleSendTelegram(ctx context.Context, t *asynq.Task) error {
	var p TelegramPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal Telegram payload: %w", err)
	}
	return h.sender.Send(p.MessageType, p.Text)
}
