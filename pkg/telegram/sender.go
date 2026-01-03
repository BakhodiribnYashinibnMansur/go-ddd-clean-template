package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) SendMessage(msgType MessageType, text string) error {
	if c.token == "" || c.chatID == "" {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)
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

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api error: status code %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) SendError(err error) error {
	return c.SendMessage(Error, fmt.Sprintf("🚨 Error: %v", err))
}

func (c *Client) SendInfo(msg string) error {
	return c.SendMessage(Info, fmt.Sprintf("ℹ️ Info: %s", msg))
}
