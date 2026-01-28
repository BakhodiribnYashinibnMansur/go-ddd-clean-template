package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrTelegramAPI is returned when Telegram API returns non-OK status.
var ErrTelegramAPI = errors.New("telegram api error")

func (c *Client) SendMessage(msgType MessageType, text string) error {
	if c.token == "" || c.chatID == "" {
		return nil
	}
	url := fmt.Sprintf(APIURLFormat, c.token)
	body := map[string]any{
		"chat_id": c.chatID,
		"text":    text,
	}

	if topicID, ok := c.topics[msgType]; ok && topicID != "" {
		body["message_thread_id"] = topicID
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, ContentTypeJSON, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
