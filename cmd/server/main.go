package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"secure-chat/internal/chat"
	"secure-chat/internal/config"
	"secure-chat/internal/logger"

	"github.com/gliderlabs/ssh"
)

func main() {
	cfg := config.Load()

	logger, err := logger.New(cfg.LogFile)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Close()

	logger.Info("Starting SSH Chat Server")
	hub := chat.NewHub(logger)

	server := &ssh.Server{
		Addr: ":" + cfg.Port,
		Handler: func(s ssh.Session) {
			// Check if PTY was requested
			ptyReq, _, isPty := s.Pty()
			if !isPty {
				io.WriteString(s, "PTY required for chat. Use: ssh -t hostname -p 8022\n")
				s.Exit(1)
				return
			}

			logger.Info("Client connected: %s (PTY: %s, Window: %dx%d)",
				s.User(), ptyReq.Term, ptyReq.Window.Width, ptyReq.Window.Height)

			client := &chat.Client{
				Stdin:  s,
				Stdout: s,
				Hub:    hub,
				Name:   s.User(),
			}

			// Get window change channel
			_, winch, _ := s.Pty()

			// Handle window changes in background
			go func() {
				for range winch {
					// Window resized
				}
			}()

			// Start client (this blocks until disconnect)
			client.Start()
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Accept all connections (Tailscale handles auth)
			return true
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			// Accept all connections (Tailscale handles auth)
			return true
		},
	}

	// Set host key
	server.SetOption(ssh.HostKeyFile("test_host_key"))

	logger.Info("SSH Chat server listening on :%s (Tailscale secure)", cfg.Port)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		logger.Info("Shutting down server...")
		cancel()
	}()

	// Start server
	go func() {
		if cfg.BindAddress != "" {
			listener, err := net.Listen("tcp", cfg.BindAddress+":"+cfg.Port)
			if err != nil {
				logger.Error("Failed to create listener: %v", err)
				return
			}
			if err := server.Serve(listener); err != nil {
				logger.Error("Server error: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil {
				logger.Error("Server error: %v", err)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("Server shutdown completed")
}
