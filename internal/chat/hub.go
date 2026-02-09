package chat

import (
	"fmt"
	"strings"
	"sync"

	"secure-chat/internal/logger"
)

type Hub struct {
	// Mutex protects the map from concurrent access hacks
	mu sync.Mutex
	// A map of connected clients to their Usernames
	clients map[*Client]string
	// The structured logger
	logger *logger.Logger
}

func NewHub(logger *logger.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]string),
		logger:  logger,
	}
}

// NOTE: Broadcast sends a message to everyone AND write the file
func (h *Hub) Broadcast(msg string, sender *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Extract username from message format "[username]: message"
	// or use sender's name directly if available
	user := "system"
	messageContent := msg

	if sender != nil {
		user = sender.Name
		// If message already has format "[username]: message", extract just the message part
		if len(msg) > 0 && msg[0] == '[' {
			if idx := strings.Index(msg, "]: "); idx != -1 {
				messageContent = msg[idx+3:]
			}
		} else {
			messageContent = msg
		}
	}

	h.logger.Chat(user, messageContent)

	for client := range h.clients {
		if client == sender {
			continue
		}
		// If sending to a specific user fails, we kick them out.
		if err := client.Write(msg + "\n"); err != nil {
			fmt.Printf("Network error for %s: %v. Disconnecting.\n", client.Name, err)

			// NOTE: We are already inside a Lock, so we must NOT call Unregister
			// (which tries to Lock again -> Deadlock).
			// We delete directly from the map.
			delete(h.clients, client)

		}
	}
}

func (h *Hub) Register(c *Client, name string) {
	h.mu.Lock()
	h.clients[c] = name
	h.mu.Unlock()
	h.Broadcast(">>> "+name+" joined.", nil)
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if name, ok := h.clients[c]; ok {
		delete(h.clients, c)
		// NOTE: calling Broadcast inside a Lock is risky in high-scale systems,
		// but fine for <100 users.
		go h.Broadcast("<<< "+name+" left.", nil)
	}
}

func (h *Hub) GetUserList() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	users := make([]string, 0, len(h.clients))
	for _, name := range h.clients {
		users = append(users, name)
	}
	return users
}
