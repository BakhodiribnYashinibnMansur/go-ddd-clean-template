package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"
	"gct/internal/kernel/infrastructure/circuitbreaker"
)

const (
	FCM_TYPE_CLIENT    = "CLIENT"
	FCM_TYPE_ADMIN     = "ADMIN"
	FCM_TYPE_CRAFTSMAN = "CRAFTSMAN"
)

const redisFCMFallbackKey = "fcm:fallback:buffer"

type fcmFallbackEntry struct {
	Token     string            `json:"token"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data,omitempty"`
	FCMType   string            `json:"fcm_type"`
	Timestamp string            `json:"timestamp"`
}

func (f *Firebase) bufferFCMFallback(token, title, body string, data map[string]string, fcmType string) error {
	if f.rdb == nil {
		f.logger.Warn("FCM circuit open and Redis unavailable, notification dropped")
		return nil
	}
	entry := fcmFallbackEntry{
		Token: token, Title: title, Body: body,
		Data: data, FCMType: fcmType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	raw, _ := json.Marshal(entry)
	// bufferFCMFallback is reached via circuit-breaker fallback paths whose
	// signatures do not carry a context, so we use Background to guarantee the
	// fallback buffer write is never cancelled by an upstream request.
	if err := f.rdb.LPush(context.Background(), redisFCMFallbackKey, raw).Err(); err != nil {
		f.logger.Warnw("FCM fallback: Redis LPUSH failed", "error", err)
		return nil
	}
	return nil
}

func (f *Firebase) selectCBAndClient(fcmType string) (*circuitbreaker.Breaker, *messaging.Client, bool) {
	switch fcmType {
	case FCM_TYPE_CLIENT:
		return f.mobileCB, f.MobileClient, true
	case FCM_TYPE_ADMIN:
		return f.webCB, f.WebClient, true
	case FCM_TYPE_CRAFTSMAN:
		return f.webCB, f.WebClient, true
	default:
		f.logger.Warnw("Unknown FCM type", "type", fcmType)
		return nil, nil, false
	}
}

func (f *Firebase) SendNotification(ctx context.Context, token, fcmType string, content Content, data map[string]string) error {
	notification := &messaging.Message{
		Token: token,
		Data:  data,
		Notification: &messaging.Notification{
			Title: content.Title,
			Body:  content.Body,
		},
	}

	cb, client, ok := f.selectCBAndClient(fcmType)
	if !ok {
		return nil
	}

	if err := cb.ExecuteWithFallback(
		func() error {
			_, err := client.Send(ctx, notification)
			if err != nil {
				f.logger.Error("Error sending notification: ", err)
				return fmt.Errorf("firebase.messaging.Send: %w", err)
			}
			return nil
		},
		func() error {
			return f.bufferFCMFallback(token, content.Title, content.Body, data, fcmType)
		},
	); err != nil {
		return fmt.Errorf("firebase.SendNotification: %w", err)
	}
	return nil
}

func (f *Firebase) SendMultiNotification(ctx context.Context, tokens []string, fcmType string, content Content, data map[string]string) error {
	validTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			validTokens = append(validTokens, token)
		}
	}
	if len(validTokens) == 0 {
		return nil
	}

	notification := &messaging.MulticastMessage{
		Data:   data,
		Tokens: validTokens,
		Notification: &messaging.Notification{
			Title: content.Title,
			Body:  content.Body,
		},
	}

	cb, client, ok := f.selectCBAndClient(fcmType)
	if !ok {
		return nil
	}

	if err := cb.ExecuteWithFallback(
		func() error {
			_, err := client.SendEachForMulticast(ctx, notification)
			if err != nil {
				f.logger.Error("Error sending multicast notification: ", err)
				return fmt.Errorf("firebase.messaging.SendEachForMulticast: %w", err)
			}
			return nil
		},
		func() error {
			for _, token := range validTokens {
				f.bufferFCMFallback(token, content.Title, content.Body, data, fcmType)
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("firebase.SendMultiNotification: %w", err)
	}
	return nil
}

func (f *Firebase) SendNotifications(ctx context.Context, tokens []string, fcmType string, content Content) error {
	cb, client, ok := f.selectCBAndClient(fcmType)
	if !ok {
		return nil
	}

	failCount := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		notification := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: content.Title,
				Body:  content.Body,
			},
		}
		err := cb.ExecuteWithFallback(
			func() error {
				_, err := client.Send(ctx, notification)
				if err != nil {
					return fmt.Errorf("firebase.messaging.Send: %w", err)
				}
				return nil
			},
			func() error {
				return f.bufferFCMFallback(token, content.Title, content.Body, nil, fcmType)
			},
		)
		if err != nil {
			failCount++
			f.logger.Warnw("Failed to send notification, skipping token",
				"token", token, "error", err)
		}
	}
	if failCount > 0 {
		f.logger.Warnw("SendNotifications completed with failures",
			"fail_count", failCount, "total", len(tokens))
	}
	return nil
}
