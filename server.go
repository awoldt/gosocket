package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Dev  AppConfig `yaml:"dev"`
	Prod AppConfig `yaml:"prod"`
}

type AppConfig struct {
	AllowedOrigins  []string `yaml:"allowed_origins"`
	WriteBufferSize int      `yaml:"write_buffer_size"`
	ReadBufferSize  int      `yaml:"read_buffer_size"`
}

var (
	mu    sync.RWMutex
	rooms = make(map[string][]*websocket.Conn) // KEY ROOM -> VALUE CONNECTIONS
)

func main() {
	cmd := &cli.Command{
		Name:        "gosocket",
		Description: "A lightweight Go-based CLI for interacting with WebSocket APIs",
		Flags: []cli.Flag{&cli.StringFlag{
			Name:  "port",
			Value: "8080",
		}},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Starts the websocket server",
				Commands: []*cli.Command{
					{
						Name:  "dev",
						Usage: "Starts the websocket server in development mode",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							port := cmd.String("port")
							config, err := initConfig("dev")
							if err != nil {
								logrus.Error("there was an error while initializing config")
								return fmt.Errorf("%s", err.Error())
							}

							logrus.SetFormatter(&logrus.TextFormatter{
								FullTimestamp:   true,
								TimestampFormat: "2006-01-02 15:04:05",
							})

							var upgrader = websocket.Upgrader{
								ReadBufferSize:  config.ReadBufferSize,
								WriteBufferSize: config.WriteBufferSize,
								CheckOrigin: func(r *http.Request) bool {
									// if no origins set, allow all
									if len(config.AllowedOrigins) == 0 {
										return true
									}
									if slices.Contains(config.AllowedOrigins, r.Host) {
										return true
									}
									return false
								},
							}

							http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
								// count the number of rooms and all connections
								htmlPage := buildStatsPage()

								w.Write([]byte(htmlPage))
							})

							http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
								conn, err := upgrader.Upgrade(w, r, nil)
								if err != nil {
									logrus.Error(err)
									return
								}

								defer conn.Close()

								// join room
								roomName := r.URL.Path
								mu.Lock()
								usersInRoom := rooms[roomName]
								usersInRoom = append(usersInRoom, conn)
								rooms[roomName] = usersInRoom
								mu.Unlock()

								logrus.Info("someone has joined room " + roomName)

								for {
									messageType, p, err := conn.ReadMessage()
									if err != nil {
										// leave room
										updatedUsersInRoom := []*websocket.Conn{}

										mu.Lock()
										for _, v := range rooms[roomName] {
											if v == conn {
												continue
											}
											updatedUsersInRoom = append(updatedUsersInRoom, v)
										}

										if len(updatedUsersInRoom) == 0 {
											// nobody in room anymore... just delete room from map
											delete(rooms, roomName)
										} else {
											rooms[roomName] = updatedUsersInRoom
										}
										mu.Unlock()

										logrus.Info("someone has left room " + roomName)
										return
									}

									// send message back to all clients within this room
									// DONT lock while WriteMessage is going (can cause bad performance for slow clients)
									// lock, create a copy of connections slice, unlock, THEN WriteMessage
									mu.RLock()
									conns := append(make([]*websocket.Conn, 0), rooms[roomName]...)
									mu.RUnlock()
									for _, v := range conns {
										if err := v.WriteMessage(messageType, p); err != nil {
											logrus.Error(err)
											return
										}
									}

								}
							})

							log.Printf("websocket server listening on :%v\n", port)
							logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func initConfig(mode string) (AppConfig, error) {
	configFile := "config.yaml"

	// see if the config file exists
	// if not, create a default one for user
	_, err := os.Stat(configFile)
	if err != nil {
		var defaultConfig Config = Config{
			Dev: AppConfig{
				AllowedOrigins:  []string{},
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			},
			Prod: AppConfig{
				AllowedOrigins:  []string{},
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			},
		}

		yamlTxt, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return AppConfig{}, fmt.Errorf("there was an error while marshalling default yaml config")
		}
		os.WriteFile(configFile, yamlTxt, 0644)
		fmt.Println("initialzed defualt config yaml")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return AppConfig{}, fmt.Errorf("there was an error while reading config file")
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return AppConfig{}, fmt.Errorf("error while reading config file")
	}

	// return the config for the mode passed in
	switch mode {
	case "dev":
		return config.Dev, nil
	case "prod":
		return config.Prod, nil
	default:
		return AppConfig{}, fmt.Errorf("%s is not a valid mode", mode)
	}
}

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

	roomListHTML := ""
	for name, conns := range rooms {
		roomListHTML += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
            </tr>`, name, len(conns))
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
`, numOfRooms, numOfConnections, roomListHTML, "Live Updates")
}
