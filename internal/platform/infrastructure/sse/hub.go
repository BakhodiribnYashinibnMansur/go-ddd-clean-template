package sse

import (
	"sync"
)

// Message represents a single SSE event to push to clients.
type Message struct {
	ID    string // Redis stream ID, used as Last-Event-ID
	Event string // SSE event type (notification, audit, monitoring, job_progress)
	Data  []byte // JSON payload
}

// Hub manages SSE client connections grouped by channel.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]map[chan Message]bool
	bufferSize int
}

// NewHub creates a new SSE hub with the given per-client buffer size.
func NewHub(bufferSize int) *Hub {
	return &Hub{
		clients:    make(map[string]map[chan Message]bool),
		bufferSize: bufferSize,
	}
}

// Register creates a new client channel for the given stream channel.
func (h *Hub) Register(channel string) chan Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Message, h.bufferSize)
	if h.clients[channel] == nil {
		h.clients[channel] = make(map[chan Message]bool)
	}
	h.clients[channel][ch] = true
	return ch
}

// Unregister removes a client channel and closes it.
func (h *Hub) Unregister(channel string, ch chan Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[channel]; ok {
		if _, exists := clients[ch]; exists {
			delete(clients, ch)
			close(ch)
		}
		if len(clients) == 0 {
			delete(h.clients, channel)
		}
	}
}

// Broadcast sends a message to all clients subscribed to a channel.
// Slow clients that have a full buffer are skipped (non-blocking send).
func (h *Hub) Broadcast(channel string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients[channel] {
		select {
		case ch <- msg:
		default:
			// skip slow client
		}
	}
}

// ClientCount returns the number of active clients on a channel.
func (h *Hub) ClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[channel])
}
