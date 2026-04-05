package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type FCMSender interface {
	Send(ctx context.Context, token, fcmType, title, body string, data map[string]string) error
	SendMulti(ctx context.Context, tokens []string, fcmType, title, body string, data map[string]string) error
}

type FCMPayload struct {
	Token   string            `json:"token"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
	FCMType string            `json:"fcm_type"`
}

type FCMMultiPayload struct {
	Tokens  []string          `json:"tokens"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
	FCMType string            `json:"fcm_type"`
}

type FCMHandler struct {
	sender FCMSender
}

func NewFCMHandler(sender FCMSender) *FCMHandler {
	return &FCMHandler{sender: sender}
}

func (h *FCMHandler) HandleSendFCM(ctx context.Context, t *asynq.Task) error {
	var p FCMPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal FCM payload: %w", err)
	}
	if err := h.sender.Send(ctx, p.Token, p.FCMType, p.Title, p.Body, p.Data); err != nil {
		return fmt.Errorf("fcm.Handler.Send: %w", err)
	}
	return nil
}

func (h *FCMHandler) HandleSendFCMMulti(ctx context.Context, t *asynq.Task) error {
	var p FCMMultiPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal FCM multi payload: %w", err)
	}
	if err := h.sender.SendMulti(ctx, p.Tokens, p.FCMType, p.Title, p.Body, p.Data); err != nil {
		return fmt.Errorf("fcm.Handler.SendMulti: %w", err)
	}
	return nil
}
