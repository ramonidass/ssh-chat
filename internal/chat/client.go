package chat

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/term"
)

type Client struct {
	Stdin  io.Reader
	Stdout io.Writer
	Hub    *Hub
	Name   string
}

// NOTE: readWriter combines an io.Reader and io.Writer into an io.ReadWriter
type readWriter struct {
	io.Reader
	io.Writer
}

// NOTE: Start begins the read loop. This blocks until the user disconnects.
func (c *Client) Start() {
	c.Hub.Register(c, c.Name)
	defer c.Hub.Unregister(c)
	welcomeMsg := "💬 Encrypted and Private Chat Server /list"

	if _, err := fmt.Fprintf(c.Stdout, "%s", welcomeMsg); err != nil {
		fmt.Printf("Failed to print welcome message: %v\n", err)
	}

	// Show recent chat history
	c.showRecentMessages(50)

	// Create a terminal for PTY-based input handling
	// This is required when PTY mode is enabled in gliderlabs/ssh
	rw := &readWriter{Reader: c.Stdin, Writer: c.Stdout}
	terminal := term.NewTerminal(rw, "> ")

	for {
		text, err := terminal.ReadLine()
		if err != nil {
			break
		}

		if len(text) > 0 && text[0] == '/' {
			c.handleCommand(text)
			continue
		}

		// Skip empty messages
		if text == "" {
			continue
		}

		// Format and broadcast message
		formatted := fmt.Sprintf("[%s]: %s", c.Name, text)
		c.Hub.Broadcast(formatted, c)
	}
}

func (c *Client) handleCommand(cmd string) {
	switch cmd {
	case "/help":
		c.showHelp()
	case "/list":
		c.listUsers()
	case "/quit":
		fmt.Fprintf(c.Stdout, "👋 Goodbye! Disconnecting...\n")
		time.Sleep(100 * time.Millisecond) // Give time for message to send
		return
	default:
		fmt.Fprintf(c.Stdout, "❓ Unknown command: %s. Type /help for available commands.\n", cmd)
	}
}

func (c *Client) showHelp() {
	help := `
Available commands:
  /help     - Show this help message
  /list     - Show online users
  /quit     - Disconnect from chat

Just type your message and press Enter to send to everyone.
`
	fmt.Fprintf(c.Stdout, help)
}

func (c *Client) listUsers() {
	users := c.Hub.GetUserList()
	userList := fmt.Sprintf("\n👥 Online users (%d):\n", len(users))
	for i, user := range users {
		userList += fmt.Sprintf("  %d. %s\n", i+1, user)
	}
	userList += "\n"
	fmt.Fprintf(c.Stdout, userList)
}

func (c *Client) Write(msg string) error {
	_, err := fmt.Fprintf(c.Stdout, "%s\n", msg)
	return err
}

// showRecentMessages displays the last N messages from chat history
func (c *Client) showRecentMessages(count int) {
	messages, err := c.Hub.logger.GetRecentMessages(count)
	if err != nil {
		// If we can't read history, just continue silently
		return
	}

	if len(messages) == 0 {
		return
	}

	fmt.Fprintf(c.Stdout, "\n📜 Recent Messages (last %d):\n", len(messages))

	for _, msg := range messages {
		// Parse the log line format: [timestamp] CHAT: user: message
		// Convert to display format: [timestamp] user: message
		displayMsg := msg
		if idx := strings.Index(msg, "CHAT:"); idx != -1 {
			displayMsg = msg[:idx] + msg[idx+6:] // Remove "CHAT: "
		}
		fmt.Fprintf(c.Stdout, "%s\n", displayMsg)
	}
}
