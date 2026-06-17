A basic WebSocket server built for smaller projects and rapid prototyping. Run one command and have a Websocket server running in seconds.

## Installation

```bash
go install github.com/awoldt/gosocket@latest
```

Or clone and build from source:

```bash
git clone https://github.com/awoldt/gosocket.git
cd gosocket
go build -o gosocket .
```

## Features

Designed for developers who need a reliable WebSocket layer without the overhead of complex message brokers.

- **Room-Based** — Clients join rooms via URL path. Messages broadcast to all room members.
- **CLI Configuration** — Configure port, origins, buffer sizes, auth, and logging via flags.
- **Token Authentication** — Optional auth token support via query parameter for secure connections.
- **Origin Control** — Configure allowed origins for CORS protection or allow all.
- **Single Binary** — Easy to deploy and scale. Just run the binary and you're live.

## Quick Start

Start the WebSocket server:

```bash
gosocket start
```

## Configuration

All settings are passed as flags on the `start` command:

| Flag | Description | Default |
| --- | --- | --- |
| `--port` | Port the server listens on | `8080` |
| `--allowed_origins` | Origins allowed to connect. Repeat the flag for multiple values. Omit to allow all. | (none) |
| `--read_buffer_size` | WebSocket read buffer size in bytes | `1024` |
| `--write_buffer_size` | WebSocket write buffer size in bytes | `1024` |
| `--auth_token` | Token required for client connections. Omit to disable auth. | (none) |
| `--logs` | Enables server-side logging | `false` |

Examples:

```bash
# Custom port with logging enabled
gosocket start --port 9000 --logs

# Restrict origins
gosocket start --port 8080 --allowed_origins localhost:3000 --allowed_origins localhost:5173

# Require authentication
gosocket start --port 8080 --auth_token my-secret-token
```

## Connecting to Rooms

Clients connect to rooms via the URL path. The path becomes the room name:

```
// Connect to the "chat" room
ws://localhost:8080/chat

// Connect to the "notifications" room
ws://localhost:8080/notifications

// Connect with authentication token
ws://localhost:8080/chat?token=your-secret-token
```

### JavaScript Example

```javascript
// Connect to a room
const ws = new WebSocket('ws://localhost:8080/chat');

ws.onopen = () => {
    console.log('Connected to chat room');
    ws.send('Hello, room!');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};

ws.onclose = () => {
    console.log('Disconnected');
};
```
