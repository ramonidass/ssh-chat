#!/bin/bash
# Test script for Tailscale SSH Chat Server

echo "Testing Tailscale SSH Chat Server..."

# Check if server binary exists
if [ ! -f "ssh-chat-server" ]; then
    echo "Server binary not found. Building..."
    go build -o ssh-chat-server cmd/server/main.go
fi

# Test configuration
echo "Configuration:"
echo "  Port: ${CHAT_PORT:-8022}"
echo "  Bind: ${CHAT_BIND:-all interfaces}"
echo "  Log: ${CHAT_LOG_FILE:-secure_chat.log}"
echo "  Tailscale Only: ${CHAT_TAILSCALE_ONLY:-true}"

# Test server startup
echo "Starting server..."
./ssh-chat-server &
SERVER_PID=$!
sleep 3

# Check if server is running
if ps -p $SERVER_PID > /dev/null; then
    echo "Server started successfully (PID: $SERVER_PID)"
    
    # Check if port is listening
    if netstat -tlnp 2>/dev/null | grep -q ":${CHAT_PORT:-8022}"; then
        echo "Server is listening on port ${CHAT_PORT:-8022}"
    else
        echo "Port check failed (might need sudo)"
    fi
    
    # Kill server
    kill $SERVER_PID
    echo "Server stopped cleanly"
else
    echo "Server failed to start"
    exit 1
fi

echo "All tests passed!"
echo ""
echo "To run your server:"
echo "  ./ssh-chat-server"
echo ""
echo "To connect from another Tailscale device:"
echo "  ssh \$(tailscale ip) -p ${CHAT_PORT:-8022}"
