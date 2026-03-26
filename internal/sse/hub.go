package sse

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	NotificationID uint   `json:"notification_id"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	Recipient      string `json:"recipient"`
	Timestamp      string `json:"timestamp"`
}

type Client struct {
	ConnectionID string
	Messages     chan *Message
	Done         chan struct{}
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) Register(connectionID string) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.clients[connectionID]; ok {
		close(existing.Done)
		delete(h.clients, connectionID)
	}

	client := &Client{
		ConnectionID: connectionID,
		Messages:     make(chan *Message, 64),
		Done:         make(chan struct{}),
	}
	h.clients[connectionID] = client
	return client
}

func (h *Hub) Unregister(connectionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[connectionID]; ok {
		close(client.Done)
		delete(h.clients, connectionID)
	}
}

func (h *Hub) Send(connectionID string, msg *Message) error {
	h.mu.RLock()
	client, ok := h.clients[connectionID]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client %s is not connected", connectionID)
	}

	msg.Timestamp = time.Now().UTC().Format(time.RFC3339)

	select {
	case client.Messages <- msg:
		return nil
	default:
		return fmt.Errorf("client %s message buffer is full", connectionID)
	}
}

func (h *Hub) IsConnected(connectionID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[connectionID]
	return ok
}
