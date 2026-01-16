# gosocket

A lightweight pub/sub WebSocket server written in Go.

## Features

- Room-based messaging via URL paths
- YAML configuration
- Optional token authentication
- Built-in stats endpoint
- Auto-generates default config on first run

## Installation

```bash
go install github.com/your-username/gosocket@latest
```

## Usage

Start the server:

```bash
gosocket start
```

Connect to a room:

```javascript
const ws = new WebSocket('ws://localhost:8080/chat');

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};

ws.send('Hello, room!');
```

## Configuration

Edit `config.yaml` to customize:

```yaml
allowed_origins: []
read_buffer_size: 1024
write_buffer_size: 1024
port: "8080"
auth_token: ""
```

## How It Works

- Clients connect to a path (e.g., `/chat`)
- The path determines the room
- Messages are broadcast to all clients in that room
- Empty rooms are automatically cleaned up

## License

MIT
