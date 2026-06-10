A basic pub/sub WebSocket server built for smaller projects and rapid prototyping. Spin up instantly, deploy with ease, and focus on your application logic, not infrastructure.

## Installation

Install gosocket using Go:

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

- **Room-Based Pub/Sub** — Clients join rooms via URL path. Messages broadcast to all room members.
- **YAML Configuration** — Simple config file for origins, buffer sizes, auth, and port settings.
- **Token Authentication** — Optional auth token support via query parameter for secure connections.
- **Origin Control** — Configure allowed origins for CORS protection or allow all.
- **Zero Config Start** — Auto-generates default config.yaml on first run. Start in seconds.
- **Single Binary** — Easy to deploy and scale. Just run the binary and you're live.

## Quick Start

Start the WebSocket server:

```bash
gosocket start
```

On first run, a `config.yaml` file is automatically created with default settings in your user config directory. The server listens on port 8080 by default.

## Configuration

gosocket saves its configuration globally. You can find and edit your `config.yaml` at:

- **Linux/macOS:** `~/.config/gosocket/config.yaml`
- **Windows:** `%AppData%\gosocket\config.yaml`

```yaml
# config.yaml
allowed_origins: []        # Empty allows all origins
read_buffer_size: 1024
write_buffer_size: 1024
port: "8080"
auth_token: ""             # Set a token to require authentication
```

| Option              | Description                                               | Default |
| ------------------- | --------------------------------------------------------- | ------- |
| allowed_origins     | List of allowed origins for CORS. Empty array allows all. | `[]`    |
| read_buffer_size    | WebSocket read buffer size in bytes.                      | 1024    |
| write_buffer_size   | WebSocket write buffer size in bytes.                     | 1024    |
| port                | Port the server listens on.                               | `"8080"` |
| auth_token          | Optional token for authentication. Empty disables auth.   | `""`    |
| logging             | Enables or disables server-side logging to the console.   | true    |

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

gosocket uses a simple pub/sub model:

1. Client connects to a WebSocket endpoint with a path (e.g., `/chat`)
2. The path determines which "room" the client joins
3. When a client sends a message, it's broadcast to **all clients in that room**
4. When a client disconnects, they're removed from the room
5. Empty rooms are automatically cleaned up

## Authentication

To enable authentication, set the `auth_token` in your config:

```yaml
auth_token: "my-secret-token"
```

Clients must then include the token as a query parameter:

```
ws://localhost:8080/chat?token=my-secret-token
```

Connections without a valid token will receive a `401 Unauthorized` response.

