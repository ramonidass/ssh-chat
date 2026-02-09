# SSH Chat Server

A secure, terminal-based chat server that uses SSH for connections. Built with Go, it provides a simple, no-client-required chat experience where users connect directly via their SSH client.

## Features

- **SSH-native** - No special client needed, just use `ssh` command
- **Real-time messaging** - Instant message broadcasting to all connected users
- **Persistent chat history** - View last 50 messages when joining (configurable)
- **User presence** - See who's online with `/list` command
- **Join/leave notifications** - Know when users enter or leave the chat
- **Secure by default** - Works seamlessly with Tailscale for zero-trust networking
- **Configurable** - Environment variables for port, bind address, and logging
- **Graceful shutdown** - Clean disconnection handling
- **Lightweight** - Single binary, minimal dependencies

## Self-Hosting

### Option 1: Local Installation

#### Prerequisites
- Go 1.21 or higher
- SSH client (for connecting)

#### Build & Run

```bash

git clone https://github.com/ramonidass/ssh-chat.git
cd ssh-chat

# Build the server
go build -o ssh-chat-server cmd/server/main.go

# Generate SSH host key (first time only)
ssh-keygen -t rsa -f test_host_key -N ""

# Run with default settings (port 8022)
./ssh-chat-server

# Or with custom configuration
CHAT_PORT=9876 ./ssh-chat-server
```

### Option 2: Docker

#### Using Docker

```bash
# Build the image
docker build -t ssh-chat-server .

# Run the container
docker run -d \
  --name ssh-chat \
  -p 8022:8022 \
  -v $(pwd)/logs:/app/logs \
  -e CHAT_PORT=8022 \
  ssh-chat-server
```

#### Using Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  ssh-chat:
    build: .
    container_name: ssh-chat
    ports:
      - "8022:8022"
    environment:
      - CHAT_PORT=8022
      - CHAT_LOG_FILE=/app/logs/chat.log
    volumes:
      - ./logs:/app/logs
      - ./test_host_key:/app/test_host_key:ro
    restart: unless-stopped
```

Then run:

```bash
# Create logs directory
mkdir -p logs

# Generate SSH host key if not exists
[ -f test_host_key ] || ssh-keygen -t rsa -f test_host_key -N ""

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAT_PORT` | `8022` | Server port |
| `CHAT_BIND` | `""` | Bind address (empty = all interfaces) |
| `CHAT_LOG_FILE` | `secure_chat.log` | Chat log file path |

### Examples

```bash
# Use custom port
CHAT_PORT=9876 ./ssh-chat-server

# Bind to specific IP only
CHAT_BIND=192.168.1.100 ./ssh-chat-server

# Custom log location
CHAT_LOG_FILE=/var/log/ssh-chat.log ./ssh-chat-server

# Disable logging
CHAT_LOG_FILE=/dev/null ./ssh-chat-server
```

## Connecting to Your Chat Server

### Connect via SSH

```bash
# Basic connection (replace with your server IP)
ssh user@your-server-ip -p 8022

# Example with custom username
ssh alice@192.168.1.100 -p 8022

# Use -t flag if you get PTY errors
ssh -t user@your-server-ip -p 8022
```

## Using the Chat

### Available Commands

- `/help` - Show available commands
- `/list` - Show online users
- `/quit` - Disconnect from chat

### Chat Features

- **Join notifications** - See when users connect (`>>> alice joined.`)
- **Leave notifications** - See when users disconnect (`<<< alice left.`)
- **Chat history** - Last 50 messages shown when you join
- **Broadcast messages** - All messages are sent to everyone

### Example Session

```
💬 Encrypted and Private Chat Server /list

📜 Recent Messages (last 3):
[2024-02-09 14:32:10] alice: Hello everyone!
[2024-02-09 14:33:22] bob: Hey alice, how are you?
[2024-02-09 14:34:05] alice: Doing great, thanks!

>>> charlie joined.
> Hello all!
[charlie]: Hello all!
[alice]: Hi charlie!
> /list

👥 Online users (3):
  1. alice
  2. bob
  3. charlie

> /quit
👋 Goodbye! Disconnecting...
```

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   SSH Client    │────▶│   SSH Chat       │────▶│   Chat Hub      │
│   (Terminal)    │     │   Server         │     │   (Broadcast)   │
│                 │     │   (Port 8022)    │     │   (Real-time)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                         │                         │
        ▼                         ▼                         ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   User Input    │◀────│   PTY Terminal   │◀────│   Message Log   │
│   (Commands)    │     │   (Read/Write)   │     │   (History)     │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Security

This server accepts any SSH key or password (PublicKeyHandler and PasswordHandler return true) because it's designed to run within a trusted network environment like:

- **Tailscale** - Recommended for zero-trust networking
- **Private LAN** - Local network behind a firewall
- **VPN** - Within a virtual private network

### Security Recommendations

```bash
# Bind to localhost only (for reverse proxy setups)
CHAT_BIND=127.0.0.1 ./ssh-chat-server

# Bind to VPN/Tailscale interface only
CHAT_BIND=100.64.0.1 ./ssh-chat-server

# Use firewall rules to restrict access
iptables -A INPUT -p tcp --dport 8022 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8022 -j DROP
```

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/ssh-chat.service`:

```ini
[Unit]
Description=SSH Chat Server
After=network.target

[Service]
Type=simple
User=chat
WorkingDirectory=/opt/ssh-chat
Environment="CHAT_PORT=8022"
Environment="CHAT_LOG_FILE=/var/log/ssh-chat/chat.log"
ExecStart=/opt/ssh-chat/ssh-chat-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable ssh-chat
sudo systemctl start ssh-chat
sudo systemctl status ssh-chat
```

### Nginx Reverse Proxy (TCP Stream)

```nginx
# /etc/nginx/nginx.conf
stream {
    server {
        listen 8022;
        proxy_pass localhost:8022;
        proxy_timeout 1s;
        proxy_connect_timeout 1s;
    }
}
```

## Troubleshooting

### Connection Issues

```bash
# Check if server is running
netstat -tlnp | grep 8022

# Test local connection
ssh localhost -p 8022

# Check server logs
tail -f secure_chat.log
```

### Common Errors

**"PTY required for chat"**
```bash
# Use the -t flag to force PTY allocation
ssh -t user@host -p 8022
```

**"no host key found"**
```bash
# Generate SSH host key
ssh-keygen -t rsa -f test_host_key -N ""
```

### Debug Mode

```bash
# Run with verbose logging
go run cmd/server/main.go
```

## License

This project is licensed under the MIT License.

## Acknowledgments

- [gliderlabs/ssh](https://github.com/gliderlabs/ssh) - Excellent SSH server library for Go
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - Terminal handling
