package sse_test

import (
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/sse"
)

func TestHub_RegisterAndBroadcast(t *testing.T) {
	hub := sse.NewHub(256)

	ch := hub.Register("notifications:user1")
	defer hub.Unregister("notifications:user1", ch)

	msg := sse.Message{
		ID:    "1234-0",
		Event: "notification",
		Data:  []byte(`{"title":"test"}`),
	}

	hub.Broadcast("notifications:user1", msg)

	select {
	case received := <-ch:
		if received.ID != "1234-0" {
			t.Errorf("expected ID '1234-0', got %q", received.ID)
		}
		if string(received.Data) != `{"title":"test"}` {
			t.Errorf("unexpected data: %s", received.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestHub_UnregisterStopsBroadcast(t *testing.T) {
	hub := sse.NewHub(256)

	ch := hub.Register("audit")
	hub.Unregister("audit", ch)

	hub.Broadcast("audit", sse.Message{ID: "1", Event: "audit", Data: []byte("test")})

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after unregister")
		}
	default:
		// channel closed, correct
	}
}

func TestHub_MultipleClients(t *testing.T) {
	hub := sse.NewHub(256)

	ch1 := hub.Register("monitoring")
	ch2 := hub.Register("monitoring")
	defer hub.Unregister("monitoring", ch1)
	defer hub.Unregister("monitoring", ch2)

	msg := sse.Message{ID: "1", Event: "error", Data: []byte("crash")}
	hub.Broadcast("monitoring", msg)

	for i, ch := range []chan sse.Message{ch1, ch2} {
		select {
		case received := <-ch:
			if received.ID != "1" {
				t.Errorf("client %d: expected ID '1', got %q", i, received.ID)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %d: timed out", i)
		}
	}
}

func TestHub_ClientCount(t *testing.T) {
	hub := sse.NewHub(256)

	if hub.ClientCount("test") != 0 {
		t.Error("expected 0 clients initially")
	}

	ch1 := hub.Register("test")
	ch2 := hub.Register("test")

	if hub.ClientCount("test") != 2 {
		t.Errorf("expected 2 clients, got %d", hub.ClientCount("test"))
	}

	hub.Unregister("test", ch1)
	if hub.ClientCount("test") != 1 {
		t.Errorf("expected 1 client after unregister, got %d", hub.ClientCount("test"))
	}

	hub.Unregister("test", ch2)
	if hub.ClientCount("test") != 0 {
		t.Errorf("expected 0 clients after all unregistered, got %d", hub.ClientCount("test"))
	}
}
