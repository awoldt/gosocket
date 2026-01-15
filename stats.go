package main

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

func buildStatsPage() string {
	type statsResponse struct {
		totalRooms       int
		totalConnections int
		rooms            map[string][]*websocket.Conn
	}

	numOfRooms := 0
	numOfConnections := 0
	mu.RLock()
	for _, v := range rooms {
		numOfRooms++
		numOfConnections += len(v)
	}

	var html strings.Builder
	for name, conns := range rooms {
		html.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
            </tr>`, name, len(conns)))
	}
	mu.RUnlock()

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Socket Server Stats</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            background-color: #f4f7f6;
        }
        .container {
            background-color: #fff;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 0.5rem;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            margin-top: 2rem;
            margin-bottom: 2rem;
        }
        .stat-card {
            background-color: #ecf0f1;
            padding: 1.5rem;
            border-radius: 6px;
            text-align: center;
        }
        .stat-value {
            display: block;
            font-size: 2.5rem;
            font-weight: bold;
            color: #3498db;
        }
        .stat-label {
            display: block;
            font-size: 1rem;
            color: #7f8c8d;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-top: 0.5rem;
        }
        .rooms-section {
            margin-top: 2rem;
        }
        table {
            min-width: 400px;
            border-collapse: collapse;
            margin-top: 1rem;
        }
        th, td {
            text-align: left;
            padding: 12px;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            color: #2c3e50;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 0.5px;
        }
        tr:hover {
            background-color: #f1f4f6;
        }
        .footer {
            margin-top: 2rem;
            font-size: 0.8rem;
            color: #bdc3c7;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Socket Server Statistics</h1>

        <div class="stats-grid">
            <div class="stat-card">
                <span class="stat-value">%d</span>
                <span class="stat-label">Active Rooms</span>
            </div>
            <div class="stat-card">
                <span class="stat-value">%d</span>
                <span class="stat-label">Total Connections</span>
            </div>
        </div>

        <div class="rooms-section">
            <h2>Active Rooms Breakdown</h2>
            <table>
                <thead>
                    <tr>
                        <th>Room Name</th>
                        <th>Connections</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>
        </div>
    </div>
    <div class="footer">
        GoSocket Server • %s
    </div>
</body>
</html>
`, numOfRooms, numOfConnections, html.String(), "Live Updates")
}
