package firebase

import (
	"context"
	"log"

	"firebase.google.com/go/v4/messaging"
)

const (
	FCM_TYPE_CLIENT    = "CLIENT"
	FCM_TYPE_ADMIN     = "ADMIN"
	FCM_TYPE_CRAFTSMAN = "CRAFTSMAN"
)

func (f *Firebase) SendNotification(ctx context.Context, token string, fcmType string, content Content, data map[string]string) error {
	notification := &messaging.Message{
		Token: token,
		Data:  data,
		Notification: &messaging.Notification{
			Title: content.Title,
			Body:  content.Body,
		},
	}
	switch fcmType {
	case FCM_TYPE_CLIENT:
		_, err := f.MobileClient.Send(ctx, notification)
		if err != nil {
			f.logger.Error("Error sending notification: ", err)
			return err
		}
	case FCM_TYPE_ADMIN:
		_, err := f.WebClient.Send(ctx, notification)
		if err != nil {
			f.logger.Error("Error sending notification: ", err)
			return err
		}
	default:
		log.Printf("Unknown FCM type: %s", fcmType)
	}

	return nil
}

func (f *Firebase) SendMultiNotification(ctx context.Context, tokens []string, fcmType string, content Content, data map[string]string) error {
	// Filter out empty tokens
	validTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			validTokens = append(validTokens, token)
		}
	}

	if len(validTokens) == 0 {
		log.Printf("No valid tokens provided")
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
	switch fcmType {
	case FCM_TYPE_CLIENT:
		_, err := f.MobileClient.SendEachForMulticast(ctx, notification)
		if err != nil {
			f.logger.Error("Error sending multicast notification: ", err)
			return err
		}
	case FCM_TYPE_CRAFTSMAN:
		_, err := f.WebClient.SendEachForMulticast(ctx, notification)
		if err != nil {
			f.logger.Error("Error sending multicast notification: ", err)
			return err
		}
	default:
		log.Printf("Unknown FCM type: %s", fcmType)
	}
	return nil
}

func (f *Firebase) SendNotifications(ctx context.Context, tokens []string, fcmType string, content Content) error {
	notification := &messaging.Message{
		Data: nil,
		Notification: &messaging.Notification{
			Title: content.Title,
			Body:  content.Body,
		},
	}
	for _, token := range tokens {
		if token == "" {
			continue
		}
		notification.Token = token
		switch fcmType {
		case FCM_TYPE_CLIENT:
			_, err := f.MobileClient.Send(ctx, notification)
			if err != nil {
				f.logger.Error("Error sending notification to token: ", err)
				return err
			}
		case FCM_TYPE_CRAFTSMAN:
			_, err := f.WebClient.Send(ctx, notification)
			if err != nil {
				f.logger.Error("Error sending notification to token: ", err)
				return err
			}
		default:
			log.Printf("Unknown FCM type: %s", fcmType)
		}
	}
	return nil
}
