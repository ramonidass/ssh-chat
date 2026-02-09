# SSH Chat Server

A secure, private chat server that leverages Tailscale's zero-trust network for authentication and encryption.

## ✨ Features

- 🔒 **Zero-trust security** - Tailscale authentication & encryption
- 🌐 **Private networking** - Only accessible via your Tailscale network
- 👥 **Real-time messaging** - Instant message broadcasting
- 📋 **User management** - See who's online with `/list`

## 🛠️ Installation

### Prerequisites
- Go 1.21+
- Tailscale client installed and running
- Your devices connected to the same Tailscale network

### Build & Run
```bash
# Clone and build
go build -o ssh-chat-server cmd/server/main.go

# Run with default settings (port 8022)
./ssh-chat-server

# Or with custom configuration
CHAT_PORT=9876 ./ssh-chat-server
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAT_PORT` | `8022` | Server port |
| `CHAT_BIND` | `""` | Bind address (empty = all interfaces) |
| `CHAT_LOG_FILE` | `secure_chat.log` | Chat log file path |
| `CHAT_TAILSCALE_ONLY` | `true` | Restrict to Tailscale interfaces only |

### Examples
```bash
# Use custom port and bind to Tailscale interface only
CHAT_PORT=9876 CHAT_BIND=100.64.0.1 ./ssh-chat-server

# Disable logging
CHAT_LOG_FILE=/dev/null ./ssh-chat-server

# Bind to specific interface (your Tailscale IP)
CHAT_BIND=100.123.45.67 ./ssh-chat-server
```

## 🔗 Connecting to Your Chat Server

### Find Your Tailscale IP
```bash
# On the server machine
tailscale ip
# Output: 100.123.45.67
```

### Connect from any Tailscale device
```bash
# Replace with your server's Tailscale IP
ssh user@100.123.45.67 -p 8022
```

### Connection Examples
```bash
# Connect as yourself (Tailscale username)
ssh 100.123.45.67 -p 8022

# Connect with a custom username (still authenticated by Tailscale)
ssh mynickname@100.123.45.67 -p 8022
```

## 💬 Using the Chat

### Available Commands
- `/help` - Show available commands
- `/list` - Show online users
- `/quit` - Disconnect from chat

## 🏗️ Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Tailscale     │────▶│   SSH Chat       │────▶│   Chat Hub      │
│   Network       │     │   Server         │     │   (Broadcast)   │
│   (Zero-trust)  │     │   (Port 8022)    │     │   (Real-time)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
       │                         │                         │
       ▼                         ▼                         ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   User Device   │◀────│   SSH Client     │◀────│   Other Users   │
│   (Any device)  │     │   (Terminal)     │     │   (Tailscale)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## 🔧 Network Configuration

### Tailscale-Only Access (Recommended)
```bash
# Bind to your Tailscale interface only
CHAT_BIND=100.64.0.1 ./ssh-chat-server
```

### Find Your Tailscale Interface
```bash
# List all interfaces
ip addr show

# Look for tailscale0 interface
# inet 100.64.0.1/32 scope global tailscale0
```

### Firewall Configuration (if needed)
```bash
# Allow Tailscale interface only (example for iptables)
iptables -A INPUT -i tailscale0 -p tcp --dport 8022 -j ACCEPT
iptables -A INPUT -p tcp --dport 8022 -j DROP
```

## 🚀 Production Deployment

### Systemd Service
```ini
# /etc/systemd/system/ssh-chat.service
[Unit]
Description=Tailscale SSH Chat Server
After=tailscaled.service
Requires=tailscaled.service

[Service]
Type=simple
User=chat
Environment="CHAT_PORT=8022"
Environment="CHAT_TAILSCALE_ONLY=true"
WorkingDirectory=/opt/ssh-chat
ExecStart=/opt/ssh-chat/ssh-chat-server
Restart=always

[Install]
WantedBy=multi-user.target
```

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ssh-chat-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ssh-chat-server .
EXPOSE 8022
CMD ["./ssh-chat-server"]
```

## 🔍 Troubleshooting

### Connection Issues
```bash
# Check if server is running
netstat -tlnp | grep 8022

# Check Tailscale connectivity
ping 100.123.45.67  # Your server's Tailscale IP

# Check server logs
tail -f secure_chat.log
```

### Permission Issues
```bash
# Ensure SSH host key exists
ssh-keygen -t rsa -f ~/.ssh/id_rsa -N ""

# Check file permissions
chmod 600 ~/.ssh/id_rsa
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Tailscale](https://tailscale.com) for zero-trust networking
- [gliderlabs/ssh](https://github.com/gliderlabs/ssh) for SSH server functionality
- Go community for excellent networking libraries
