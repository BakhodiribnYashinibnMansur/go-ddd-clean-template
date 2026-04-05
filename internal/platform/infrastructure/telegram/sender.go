package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrTelegramAPI is returned when Telegram API returns non-OK status.
var ErrTelegramAPI = errors.New("telegram api error")

func (c *Client) SendMessage(msgType MessageType, text string) error {
	if c.token == "" || c.chatID == "" {
		return nil
	}
	return c.telegramCB.ExecuteWithFallback(
		func() error {
			return c.doSend(msgType, text)
		},
		func() error {
			return c.bufferToPending(msgType, text)
		},
	)
}

func (c *Client) doSend(msgType MessageType, text string) error {
	url := fmt.Sprintf(APIURLFormat, c.token)
	body := map[string]any{
		"chat_id": c.chatID,
		"text":    text,
	}

	if topicID, ok := c.topics[msgType]; ok && topicID != "" {
		body["message_thread_id"] = topicID
	}

	resp, _, err := c.http.PostJSON(context.Background(), url, "SendMessage", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status code %d", ErrTelegramAPI, resp.StatusCode)
	}
	return nil
}

func (c *Client) SendError(err error) error {
	return c.SendMessage(Error, fmt.Sprintf("%s%v", PrefixError, err))
}

func (c *Client) SendInfo(msg string) error {
	return c.SendMessage(Info, PrefixInfo+msg)
}

const redisTelegramPendingKey = "telegram:pending"

type telegramPendingEntry struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
	FailedAt    string `json:"failed_at"`
}

func (c *Client) bufferToPending(msgType MessageType, text string) error {
	if c.rdb == nil {
		if c.log != nil {
			c.log.Warnw("Telegram circuit open and Redis unavailable, message dropped",
				"message_type", string(msgType))
		}
		return nil
	}
	entry := telegramPendingEntry{
		MessageType: string(msgType),
		Text:        text,
		FailedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := c.rdb.LPush(context.Background(), redisTelegramPendingKey, data).Err(); err != nil {
		if c.log != nil {
			c.log.Warnw("Telegram fallback: Redis LPUSH failed, message dropped", "error", err)
		}
		return nil
	}
	return nil
}
