package pubsub

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

func TestPubSub_PublishSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		channel       string
		message       string
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success publish and subscribe",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			channel:       "test_channel",
			message:       "hello world",
			expectedError: false,
		},
		{
			name: "publish to empty channel",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			channel:       "",
			message:       "test message",
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "channel")
			},
		},
		{
			name: "publish empty message",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			channel:       "test_channel",
			message:       "",
			expectedError: false,
		},
		{
			name: "redis connection error on publish",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			channel:       "test_channel",
			message:       "test message",
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "publish message with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			channel:       "test_channel",
			message:       "message with special chars: !@#$%^&*()",
			expectedError: false,
		},
		{
			name: "publish large message",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			channel:       "test_channel",
			message:       strings.Repeat("large message ", 1000),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			ps := New(client)
			testCtx := t.Context()

			// act
			err := ps.Publish(testCtx, tt.channel, tt.message)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPubSub_PublishSubscribeOriginal(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	ps := New(client)
	ctx := t.Context()
	channel := "test_channel"

	// Subscribe
	pubsub := ps.Subscribe(ctx, channel)
	defer ps.Close(pubsub)

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	require.NoError(t, err)

	msgContent := "hello world"
	done := make(chan error, 1)

	go func() {
		// Small delay to ensure subscriber is ready (though Receive should have handled it)
		time.Sleep(10 * time.Millisecond)
		err := ps.Publish(t.Context(), channel, msgContent)
		done <- err
	}()

	// Receive with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	msg, err := pubsub.ReceiveMessage(ctxTimeout)
	require.NoError(t, err)
	assert.Equal(t, channel, msg.Channel)
	assert.Equal(t, msgContent, msg.Payload)

	err = <-done
	require.NoError(t, err)
}

func TestPubSub_PSubscribe(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	ps := New(client)
	ctx := t.Context()
	pattern := "test_*"

	// PSubscribe
	pubsub := ps.PSubscribe(ctx, pattern)
	defer ps.Close(pubsub)

	// Wait for subscription
	_, err := pubsub.Receive(ctx)
	require.NoError(t, err)

	channel := "test_1"
	msgContent := "pattern match"

	done := make(chan error, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		err := ps.Publish(t.Context(), channel, msgContent)
		done <- err
	}()

	// Receive
	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	msg, err := pubsub.ReceiveMessage(ctxTimeout)
	require.NoError(t, err)
	assert.Equal(t, channel, msg.Channel)
	assert.Equal(t, msgContent, msg.Payload)

	err = <-done
	require.NoError(t, err)
}
