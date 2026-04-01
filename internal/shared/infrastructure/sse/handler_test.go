package sse

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Handler constructor tests
// ---------------------------------------------------------------------------

func TestNewHandler(t *testing.T) {
	hub := NewHub(10)
	h := NewHandler(hub, 30*time.Second)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.hub != hub {
		t.Error("expected handler to reference the provided hub")
	}
	if h.heartbeatInterval != 30*time.Second {
		t.Errorf("expected 30s heartbeat, got %v", h.heartbeatInterval)
	}
}

func TestNewHandler_CustomInterval(t *testing.T) {
	hub := NewHub(5)
	h := NewHandler(hub, 10*time.Second)
	if h.heartbeatInterval != 10*time.Second {
		t.Errorf("expected 10s interval, got %v", h.heartbeatInterval)
	}
}

// ---------------------------------------------------------------------------
// Hub unit tests
// ---------------------------------------------------------------------------

func TestHub_NewHub(t *testing.T) {
	hub := NewHub(20)
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.bufferSize != 20 {
		t.Errorf("expected buffer size 20, got %d", hub.bufferSize)
	}
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub(5)
	ch := hub.Register("test-channel")
	if hub.ClientCount("test-channel") != 1 {
		t.Error("expected 1 client after register")
	}

	hub.Unregister("test-channel", ch)
	if hub.ClientCount("test-channel") != 0 {
		t.Error("expected 0 clients after unregister")
	}
}

func TestHub_RegisterMultipleChannels(t *testing.T) {
	hub := NewHub(5)
	ch1 := hub.Register("channel-a")
	ch2 := hub.Register("channel-b")
	defer hub.Unregister("channel-a", ch1)
	defer hub.Unregister("channel-b", ch2)

	if hub.ClientCount("channel-a") != 1 {
		t.Error("expected 1 client on channel-a")
	}
	if hub.ClientCount("channel-b") != 1 {
		t.Error("expected 1 client on channel-b")
	}
}

func TestHub_BroadcastMessage(t *testing.T) {
	hub := NewHub(5)
	ch := hub.Register("chan-1")
	defer hub.Unregister("chan-1", ch)

	msg := Message{ID: "1", Event: "test", Data: []byte("hello")}
	hub.Broadcast("chan-1", msg)

	select {
	case received := <-ch:
		if received.ID != "1" {
			t.Errorf("expected ID '1', got %q", received.ID)
		}
		if received.Event != "test" {
			t.Errorf("expected event 'test', got %q", received.Event)
		}
		if string(received.Data) != "hello" {
			t.Errorf("expected data 'hello', got %q", string(received.Data))
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for broadcast message")
	}
}

func TestHub_BroadcastToNoSubscribers(t *testing.T) {
	hub := NewHub(5)
	// Should not panic
	hub.Broadcast("nonexistent", Message{Event: "test", Data: []byte("data")})
}

func TestHub_BroadcastSkipsSlowClient(t *testing.T) {
	hub := NewHub(1) // buffer of 1
	ch := hub.Register("chan-1")
	defer hub.Unregister("chan-1", ch)

	// Fill the buffer
	hub.Broadcast("chan-1", Message{Event: "msg1", Data: []byte("first")})
	// This should be dropped (non-blocking)
	hub.Broadcast("chan-1", Message{Event: "msg2", Data: []byte("second")})

	select {
	case msg := <-ch:
		if msg.Event != "msg1" {
			t.Errorf("expected first message, got %q", msg.Event)
		}
	default:
		t.Error("expected at least one message in buffer")
	}

	// Verify second message was dropped
	select {
	case msg := <-ch:
		t.Errorf("expected no second message, but got event %q", msg.Event)
	default:
		// expected: nothing in buffer
	}
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub(5)
	ch1 := hub.Register("shared")
	ch2 := hub.Register("shared")
	defer hub.Unregister("shared", ch1)
	defer hub.Unregister("shared", ch2)

	if hub.ClientCount("shared") != 2 {
		t.Errorf("expected 2 clients, got %d", hub.ClientCount("shared"))
	}

	hub.Broadcast("shared", Message{Event: "ping", Data: []byte("hi")})

	for i, ch := range []chan Message{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg.Event != "ping" {
				t.Errorf("client %d: expected 'ping', got %q", i, msg.Event)
			}
			if string(msg.Data) != "hi" {
				t.Errorf("client %d: expected data 'hi', got %q", i, string(msg.Data))
			}
		case <-time.After(1 * time.Second):
			t.Errorf("client %d: timed out waiting for message", i)
		}
	}
}

func TestHub_UnregisterNonexistentChannel(t *testing.T) {
	hub := NewHub(5)
	ch := make(chan Message)
	// Should not panic
	hub.Unregister("nonexistent", ch)
}

func TestHub_UnregisterNonexistentClient(t *testing.T) {
	hub := NewHub(5)
	real := hub.Register("test")
	defer hub.Unregister("test", real)

	fake := make(chan Message)
	// Unregistering a channel that wasn't registered should not panic
	hub.Unregister("test", fake)

	// Real client should still be there
	if hub.ClientCount("test") != 1 {
		t.Error("expected real client to still be registered")
	}
}

func TestHub_UnregisterCleansUpEmptyChannel(t *testing.T) {
	hub := NewHub(5)
	ch := hub.Register("cleanup-test")
	hub.Unregister("cleanup-test", ch)

	if hub.ClientCount("cleanup-test") != 0 {
		t.Error("expected 0 after last client unregistered")
	}

	// Verify the channel entry itself is removed
	hub.mu.RLock()
	_, exists := hub.clients["cleanup-test"]
	hub.mu.RUnlock()
	if exists {
		t.Error("expected channel map entry to be removed")
	}
}

func TestHub_ClientCount_EmptyChannel(t *testing.T) {
	hub := NewHub(5)
	count := hub.ClientCount("does-not-exist")
	if count != 0 {
		t.Errorf("expected 0 for nonexistent channel, got %d", count)
	}
}

func TestHub_BroadcastIsolation(t *testing.T) {
	hub := NewHub(5)
	chA := hub.Register("channel-a")
	chB := hub.Register("channel-b")
	defer hub.Unregister("channel-a", chA)
	defer hub.Unregister("channel-b", chB)

	hub.Broadcast("channel-a", Message{Event: "only-a", Data: []byte("a-data")})

	// channel-a should have the message
	select {
	case msg := <-chA:
		if msg.Event != "only-a" {
			t.Errorf("expected 'only-a', got %q", msg.Event)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for channel-a message")
	}

	// channel-b should have nothing
	select {
	case msg := <-chB:
		t.Errorf("expected no message on channel-b, got event %q", msg.Event)
	default:
		// expected
	}
}

func TestMessage_Fields(t *testing.T) {
	msg := Message{
		ID:    "stream-id-123",
		Event: "notification",
		Data:  []byte(`{"text":"hello world"}`),
	}
	if msg.ID != "stream-id-123" {
		t.Errorf("expected ID 'stream-id-123', got %q", msg.ID)
	}
	if msg.Event != "notification" {
		t.Errorf("expected Event 'notification', got %q", msg.Event)
	}
	if string(msg.Data) != `{"text":"hello world"}` {
		t.Errorf("unexpected Data: %q", string(msg.Data))
	}
}
