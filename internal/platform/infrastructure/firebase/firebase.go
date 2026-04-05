package firebase

import (
	"context"
	"encoding/json"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"gct/config"
	"gct/internal/platform/infrastructure/circuitbreaker"
	"gct/internal/platform/infrastructure/logger"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
)

type Firebase struct {
	MobileApp    *firebase.App
	MobileClient *messaging.Client
	WebApp       *firebase.App
	WebClient    *messaging.Client
	logger       logger.Log
	rdb          redis.Cmdable
	mobileCB     *circuitbreaker.Breaker
	webCB        *circuitbreaker.Breaker
}

func NewFirebase(ctx context.Context, logger logger.Log, cfg config.Firebase, rdb redis.Cmdable) (*Firebase, error) {
	// Initialize Mobile Firebase app
	mobileCredentials, err := json.Marshal(cfg.Mobile)
	if err != nil {
		return nil, err
	}
	mobileOpt := option.WithCredentialsJSON(mobileCredentials)
	mobileApp, err := firebase.NewApp(ctx, nil, mobileOpt)
	if err != nil {
		return nil, err
	}

	// Initialize Web Firebase app
	webCredentials, err := json.Marshal(cfg.Web)
	if err != nil {
		return nil, err
	}
	webOpt := option.WithCredentialsJSON(webCredentials)
	webApp, err := firebase.NewApp(ctx, nil, webOpt)
	if err != nil {
		return nil, err
	}

	// Get FCM clients
	mobileFcmClient, err := mobileApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	webFcmClient, err := webApp.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &Firebase{
		MobileApp:    mobileApp,
		MobileClient: mobileFcmClient,
		WebApp:       webApp,
		WebClient:    webFcmClient,
		logger:       logger,
		rdb:          rdb,
		mobileCB: circuitbreaker.New("firebase-mobile", circuitbreaker.Config{
			FailureThreshold: 3,
			Timeout:          60 * time.Second,
		}),
		webCB: circuitbreaker.New("firebase-web", circuitbreaker.Config{
			FailureThreshold: 3,
			Timeout:          60 * time.Second,
		}),
	}, nil
}
